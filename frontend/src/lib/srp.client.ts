import { modPow } from 'bigint-toolkit';

export type LoginError =
  | 'INVALID_CREDENTIALS'
  | 'SERVER_ERROR'
  | 'NETWORK_ERROR';

export default class SRPClient {
  private N: bigint = 0n;
  private g: bigint = 0n;

  private I: string = '';
  private P: string = '';

  private a: bigint = 0n;
  private A: bigint = 0n;

  private s: string = '';
  private B: bigint = 0n;

  private u: bigint = 0n;
  private k: bigint = 0n;
  private x: bigint = 0n;
  private premaster: bigint = 0n;

  // should call on login page load
  public async init() {
    const data = (await this.request('/v1/srp.json', 'GET')) as {
      prime: string;
      generator: string;
    };

    this.N = BigInt(`0x${data.prime}`);
    this.g = BigInt(`0x${data.generator}`);

    this.a = this.generatePrivateKey();
    this.A = this.generatePublicKey();
  }

  public async login(
    username: string,
    password: string
  ): Promise<{ success: true } | { success: false; error: LoginError }> {
    this.I = username;
    this.P = password;

    try {
      await this.initiateAuthentication();

      this.calculatePremaster();

      await this.verify();
    } catch (error) {
      return this.handleLoginError(error);
    }

    return { success: true };
  }

  private async verify() {
    try {
      const response = (await this.request('/v1/srp/verify', 'POST', {
        premaster: this.premaster.toString(),
      })) as { success: boolean };

      if (!response.success) {
        throw new Error('Invalid challenge');
      }
    } catch (error) {
      throw error;
    }
  }

  private async calculatePremaster() {
    await this.calculateu();
    await this.calculatek();
    await this.calculatex();

    this.premaster =
      (this.B - this.k * this.g ** this.x) ** (this.a + this.u * this.x) %
      this.N;
  }

  private async calculateu() {
    const paddedA = this.A.toString(16).padStart(512, '0');
    const paddedB = this.B.toString(16).padStart(512, '0');
    const concatenated = paddedA + paddedB;
    const hash = await this.hash(concatenated);
    this.u = BigInt(`0x${hash}`);
  }

  private async calculatex() {
    const inner = await this.hash(`${this.I}:${this.P}`);
    const concatenated = this.s + inner;
    const x = await this.hash(concatenated);
    this.x = BigInt(`0x${x}`);
  }

  private async calculatek() {
    const paddedG = this.g.toString(16).padStart(512, '0');
    const nHex = this.N.toString(16);
    const k = await this.hash(nHex + paddedG);
    this.k = BigInt(`0x${k}`);
  }

  private handleLoginError(error: unknown): {
    success: false;
    error: LoginError;
  } {
    console.error('Login error:', error);

    if (error instanceof Error) {
      if (error.message.includes('Invalid challenge')) {
        return { success: false, error: 'INVALID_CREDENTIALS' };
      }
      if (
        error.message.includes('NetworkError') ||
        error.message.includes('Failed to fetch')
      ) {
        return { success: false, error: 'NETWORK_ERROR' };
      }
      return { success: false, error: 'SERVER_ERROR' };
    }

    return { success: false, error: 'SERVER_ERROR' };
  }

  private async initiateAuthentication() {
    try {
      const response = (await this.request('/v1/srp/challenge', 'POST', {
        I: this.I.toString(),
        A: this.A.toString(),
      })) as { s: string; B: string };

      if (!response.s || !response.B) {
        throw new Error('Invalid challenge');
      }

      this.s = response.s;
      this.B = BigInt(`0x${response.B}`);
    } catch (error) {
      throw error;
    }
  }
  private generatePublicKey(): bigint {
    // g^a % N
    return modPow(this.g, this.a, this.N);
  }

  private generatePrivateKey(bitLength: number = 256): bigint {
    const byteLength = bitLength / 8;
    const randomBytes = new Uint8Array(byteLength);
    crypto.getRandomValues(randomBytes);

    return BigInt(
      '0x' +
        Array.from(randomBytes, (byte) =>
          byte.toString(16).padStart(2, '0')
        ).join('')
    );
  }

  private async hash(data: any): Promise<string> {
    const encoder = new TextEncoder();
    const dataBuffer = encoder.encode(JSON.stringify(data));
    const hashBuffer = await crypto.subtle.digest('SHA-256', dataBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
  }

  private async request(
    url: string,
    method: 'GET' | 'POST',
    body?: any,
    headers: Record<string, string> = {}
  ) {
    try {
      const defaultHeaders = {
        'Content-Type': 'application/json',
        ...headers,
      };
      const config: RequestInit = {
        method,
        headers: defaultHeaders,
        ...(body && method === 'POST' && { body: JSON.stringify(body) }),
      };
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorBody = await response.text();
        throw new Error(
          `HTTP error! status: ${response.status}, message: ${errorBody}`
        );
      }
      return await response.json();
    } catch (error) {
      throw error;
    }
  }
}
