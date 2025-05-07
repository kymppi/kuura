import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios';
import type { User } from '../contexts/AuthProvider';
import {
  SRPClient,
  type PrimeField,
  type SRPClientInstance,
} from './srp.client';

export type LoginError =
  | 'INVALID_CREDENTIALS'
  | 'SERVER_ERROR'
  | 'NETWORK_ERROR'
  | 'SUSPICIOUS_SERVER'
  | 'CLIENT_UNSUPPORTED'
  | 'TOKEN_REFRESH_FAILED'
  | 'TOKEN_REFRESH_COOLDOWN';

interface AuthenticationResponse {
  data: string;
}

interface VerificationResponse {
  success: boolean;
  data: string;
}

interface TokenRefreshResponse {
  success: boolean;
}

const SKIP_REFRESH_URLS = ['/v1/srp', '/v1/logout', '/v1/user/tokens/internal'];

export class SRPAuthClient {
  private readonly srpClient: SRPClient;
  private readonly axiosInstance: AxiosInstance;
  private currentClient?: SRPClientInstance;
  private isRefreshing: boolean = false;
  private refreshSubscribers: Array<(status: string) => void> = [];
  private lastRefreshTime: number = 0;
  private readonly REFRESH_COOLDOWN = 5000;

  constructor(baseUrl: string, primeField: PrimeField) {
    this.srpClient = new SRPClient(primeField);

    this.axiosInstance = axios.create({
      baseURL: baseUrl,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.axiosInstance.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config;

        const skipRefresh = SKIP_REFRESH_URLS.some((url) =>
          originalRequest?.url?.includes(url)
        );

        if (
          !originalRequest ||
          error.response?.status !== 401 ||
          (originalRequest as any)._retry ||
          skipRefresh
        ) {
          return Promise.reject(error);
        }

        if (this.isInCooldownPeriod()) {
          return Promise.reject(new Error('TOKEN_REFRESH_COOLDOWN'));
        }

        if (this.isRefreshing) {
          try {
            await new Promise<string>((resolve, reject) => {
              this.refreshSubscribers.push((status: string) => {
                if (status === 'refreshed') {
                  resolve(status);
                } else {
                  reject('Token refresh failed');
                }
              });
            });

            (originalRequest as any)._retry = true;
            return this.axiosInstance(originalRequest);
          } catch (refreshError) {
            return Promise.reject(refreshError);
          }
        } else {
          (originalRequest as any)._retry = true;
          this.isRefreshing = true;
          this.refreshSubscribers = [];

          try {
            const refreshSuccess = await this.refreshAccessToken();

            if (refreshSuccess) {
              this.refreshSubscribers.forEach((callback) =>
                callback('refreshed')
              );

              return this.axiosInstance(originalRequest);
            }

            this.notifyRefreshFailed();
            return Promise.reject(new Error('TOKEN_REFRESH_FAILED'));
          } catch (refreshError) {
            this.notifyRefreshFailed();
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }
      }
    );
  }

  private isInCooldownPeriod(): boolean {
    const now = Date.now();
    return now - this.lastRefreshTime < this.REFRESH_COOLDOWN;
  }

  private notifyRefreshFailed(): void {
    this.refreshSubscribers.forEach((callback) => callback('failed'));
    this.refreshSubscribers = [];
  }

  public async login(
    loginTarget: string,
    username: string,
    password: string
  ): Promise<{ success: true } | { success: false; error: LoginError }> {
    if (!this.isClientSupported()) {
      return { success: false, error: 'CLIENT_UNSUPPORTED' };
    }

    try {
      // Step 1: Initialize authentication
      const authResponse = await this.initiateAuthentication(
        username,
        password
      );
      if (!authResponse.success) {
        return { success: false, error: 'INVALID_CREDENTIALS' };
      }

      // Step 2: Process server challenge
      const verifyResponse = await this.processServerChallenge(
        authResponse.data,
        loginTarget
      );
      if (!verifyResponse.success) {
        return { success: false, error: 'INVALID_CREDENTIALS' };
      }

      if (!this.currentClient)
        return { success: false, error: 'SUSPICIOUS_SERVER' };

      const serverOk = await this.srpClient.verifyServer(
        this.currentClient,
        verifyResponse.data
      );

      if (!serverOk) return { success: false, error: 'SUSPICIOUS_SERVER' };

      return { success: true };
    } catch (error) {
      return this.handleLoginError(error);
    }
  }

  private async initiateAuthentication(
    username: string,
    password: string
  ): Promise<{ success: boolean; data: string }> {
    try {
      this.currentClient = await this.srpClient.newClient(
        new TextEncoder().encode(username),
        new TextEncoder().encode(password)
      );

      const credentials = this.srpClient.getCredentials(this.currentClient);

      const response = await this.request<AuthenticationResponse>(
        '/v1/srp/begin',
        'POST',
        {
          data: credentials,
        }
      );

      if (!response?.data) {
        throw new Error('Invalid server response');
      }

      return { success: true, data: response.data };
    } catch (error) {
      console.error('Authentication initiation failed:', error);
      throw error;
    }
  }

  private async processServerChallenge(
    serverData: string,
    loginTarget: string
  ): Promise<{ success: boolean; data: string }> {
    if (!this.currentClient) return { success: false, data: '' };

    const payload = await this.srpClient.getVerifyData(
      this.currentClient,
      serverData
    );

    try {
      const response = await this.request<VerificationResponse>(
        '/v1/srp/verify',
        'POST',
        {
          identity: this.srpClient.getIdentity(this.currentClient),
          data: payload,
          target_service: loginTarget,
        }
      );

      return { success: !!response?.success, data: response.data };
    } catch (error) {
      console.error('Challenge verification failed:', error);
      throw error;
    }
  }

  private async request<T>(
    endpoint: string,
    method: 'GET' | 'POST',
    data?: any,
    config: AxiosRequestConfig = {}
  ): Promise<T> {
    try {
      const axiosConfig = { ...config };

      if (method === 'GET') {
        axiosConfig.params = data;
      }

      const response = await this.axiosInstance.request<T>({
        url: endpoint,
        method,
        data: method === 'POST' ? data : undefined,
        ...axiosConfig,
      });

      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        if (
          error.code === 'ECONNABORTED' ||
          error.message.includes('Network Error')
        ) {
          throw new Error('NETWORK_ERROR');
        }

        if (error.response?.status === 401) {
          throw new Error('Invalid credentials');
        }

        throw new Error(
          `HTTP error! status: ${error.response?.status ?? 'unknown'}`
        );
      }

      throw error;
    }
  }

  private handleLoginError(error: unknown): {
    success: false;
    error: LoginError;
  } {
    console.error('Login error:', error);
    if (error instanceof Error) {
      if (error.message === 'NETWORK_ERROR') {
        return { success: false, error: 'NETWORK_ERROR' };
      }
      if (error.message.includes('Invalid credentials')) {
        return { success: false, error: 'INVALID_CREDENTIALS' };
      }
      if (error.message === 'TOKEN_REFRESH_FAILED') {
        return { success: false, error: 'TOKEN_REFRESH_FAILED' };
      }
      if (error.message === 'TOKEN_REFRESH_COOLDOWN') {
        return { success: false, error: 'TOKEN_REFRESH_COOLDOWN' };
      }
    }

    return { success: false, error: 'SERVER_ERROR' };
  }

  public async refreshAccessToken(): Promise<boolean> {
    try {
      this.lastRefreshTime = Date.now();

      const response = await this.axiosInstance.post<TokenRefreshResponse>(
        '/v1/user/tokens/internal'
      );

      return response.data.success;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return false;
    }
  }

  public async getUser(): Promise<User | null> {
    try {
      const response = await this.axiosInstance.get<{
        id: string;
        username: string;
        last_login_at: string;
      }>('/v1/me');

      return {
        ...response.data,
        last_login_at: new Date(response.data.last_login_at),
      };
    } catch (error) {
      console.error('Failed to get authenticated user:', error);
      return null;
    }
  }

  public async logout(): Promise<boolean> {
    try {
      const response = await this.axiosInstance.post<{ success: boolean }>(
        '/v1/logout'
      );

      return response.data.success;
    } catch (error) {
      console.error('Failed to logout:', error);
      return false;
    }
  }

  private isClientSupported(): boolean {
    return (
      typeof window !== 'undefined' &&
      !!window.crypto &&
      !!window.crypto.subtle &&
      typeof TextEncoder !== 'undefined'
    );
  }
}
