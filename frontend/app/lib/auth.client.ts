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
  | 'CLIENT_UNSUPPORTED';

interface AuthenticationResponse {
  data: string;
}

interface VerificationResponse {
  success: boolean;
  data: string;
}

export class SRPAuthClient {
  private readonly baseUrl: string;
  private readonly srpClient: SRPClient;
  private currentClient?: SRPClientInstance;

  constructor(baseUrl: string, primeField: PrimeField) {
    this.baseUrl = baseUrl;
    this.srpClient = new SRPClient(primeField);
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
    body?: any,
    headers: Record<string, string> = {}
  ): Promise<T> {
    try {
      const url = `${this.baseUrl}${endpoint}`;
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers,
        },
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!response.ok) {
        if (response.status == 401) {
          throw new Error('Invalid credentials');
        }

        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      if (error instanceof Error) {
        if (error.message.includes('Failed to fetch')) {
          throw new Error('NETWORK_ERROR');
        }
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
    }

    return { success: false, error: 'SERVER_ERROR' };
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
