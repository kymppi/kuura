import { modPow } from 'bigint-toolkit';

export interface PrimeField {
  g: bigint;
  N: bigint;
  n: number;
}

export const DefaultPrimeField: PrimeField = {
  // the 4096-bit Group from RFC 5054
  g: BigInt(5),
  N: BigInt(
    '0xFFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AAAC42DAD33170D04507A33A85521ABDF1CBA64ECFB850458DBEF0A8AEA71575D060C7DB3970F85A6E1E4C7ABF5AE8CDB0933D71E8C94E04A25619DCEE3D2261AD2EE6BF12FFA06D98A0864D87602733EC86A64521F2B18177B200CBBE117577A615D6C770988C0BAD946E208E24FA074E5AB3143DB5BFCE0FD108E4B82D120A92108011A723C12A787E6D788719A10BDBA5B2699C327186AF4E23C1A946834B6150BDA2583E9CA2AD44CE8DBBBC2DB04DE8EF92E8EFC141FBECAA6287C59474E6BC05D99B2964FA090C3A2233BA186515BE7ED1F612970CEE2D7AFB81BDD762170481CD0069127D5B05AA993B4EA988D8FDDC186FFB7DC90A6C08F4DF435C934063199FFFFFFFFFFFFFFFF'
  ),
  n: 512,
};

export interface SRPClientInstance {
  s: SRPClient;
  i: Uint8Array;
  p: Uint8Array;
  a: bigint;
  xA: bigint;
  k: bigint;
  xK?: Uint8Array;
  xM?: Uint8Array;
}

export class SRPClient {
  private readonly pf: PrimeField;
  private readonly h: string = 'SHA-256';

  constructor(primeField: PrimeField) {
    this.pf = primeField;
  }

  /**
   * Creates a new SRP client instance
   */
  async newClient(I: Uint8Array, p: Uint8Array): Promise<SRPClientInstance> {
    const i = await this.hashbyte(I);
    const hashedPassword = await this.hashbyte(p);
    const a = await this.randBigInt(this.pf.n * 8);

    // Calculate k = H(N, pad(g))
    const paddedG = this.pad(this.bigIntToUint8Array(this.pf.g), this.pf.n);
    const k = await this.hashbigint(
      this.bigIntToUint8Array(this.pf.N),
      paddedG
    );

    // Calculate A = g^a % N
    const xA = modPow(this.pf.g, a, this.pf.N);

    return {
      s: this,
      i,
      p: hashedPassword,
      a,
      xA,
      k,
    };
  }

  /**
   * Returns the hashed identity
   */
  getIdentity(client: SRPClientInstance): string {
    const iHex = Array.from(client.i)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');

    return iHex;
  }

  /**
   * Returns client credentials in the same format as the Go version
   */
  getCredentials(client: SRPClientInstance): string {
    const iHex = Array.from(client.i)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');

    const xAHex = client.xA.toString(16);

    return `${iHex}:${xAHex}`;
  }

  /**
   * Validates server public credentials and generates session key
   * Returns the mutual authenticator
   */
  async getVerifyData(
    client: SRPClientInstance,
    serverData: string
  ): Promise<string> {
    const [saltHex, BHex] = serverData.split(':');
    if (!saltHex || !BHex) {
      throw new Error('srp: invalid server public key');
    }

    const salt = this.hexToUint8Array(saltHex);
    if (!salt) {
      throw new Error('srp: invalid server public key');
    }

    const B = BigInt(`0x${BHex}`);
    if (B === 0n) {
      throw new Error('srp: invalid server public key');
    }

    // Verify B % N != 0
    if (B % this.pf.N === 0n) {
      throw new Error('srp: invalid server public key');
    }

    // Calculate u = H(pad(A), pad(B))
    const paddedA = this.pad(this.bigIntToUint8Array(client.xA), this.pf.n);
    const paddedB = this.pad(this.bigIntToUint8Array(B), this.pf.n);
    const u = await this.hashbigint(paddedA, paddedB);
    if (u === 0n) {
      throw new Error('srp: invalid server public key');
    }

    // Calculate x = H(i, p, salt)
    const x = await this.hashbigint(client.i, client.p, salt);

    // Calculate S = ((B - kg^x) ^ (a + ux)) % N
    const gx = modPow(this.pf.g, x, this.pf.N);
    const kgx = (client.k * gx) % this.pf.N;
    const t1 = (B + this.pf.N - kgx) % this.pf.N; // Use + N - kgx instead of direct subtraction
    const t2 = (client.a + u * x) % (this.pf.N - 1n); // Reduce exponent modulo (N-1)
    const S = modPow(t1, t2, this.pf.N); // even though the math changed a bit, S remains the same

    // Calculate K = H(S)
    client.xK = await this.hashbyte(this.bigIntToUint8Array(S));

    // Calculate M = H(K, A, B, i, salt, N, g)
    client.xM = await this.hashbyte(
      client.xK,
      this.bigIntToUint8Array(client.xA),
      this.bigIntToUint8Array(B),
      client.i,
      salt,
      this.bigIntToUint8Array(this.pf.N),
      this.bigIntToUint8Array(this.pf.g)
    );

    // Return M as hex string
    return Array.from(client.xM)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');
  }

  async verifyServer(
    client: SRPClientInstance,
    proof: string
  ): Promise<boolean> {
    if (!client.xK) return false;
    if (!client.xM) return false;

    const hash = await this.hashbyte(client.xK, client.xM);

    const proofBytes = this.hexToUint8Array(proof);

    return this.constantTimeEqual(proofBytes, hash);
  }

  /**
   * Generates a random bigint of specified bits
   */
  private async randBigInt(bits: number): Promise<bigint> {
    const bytes = Math.ceil(bits / 8);
    const array = new Uint8Array(bytes);
    crypto.getRandomValues(array);
    return BigInt(
      '0x' +
        Array.from(array)
          .map((b) => b.toString(16).padStart(2, '0'))
          .join('')
    );
  }

  /**
   * Pads a Uint8Array to a specified length
   */
  private pad(input: Uint8Array, length: number): Uint8Array {
    const result = new Uint8Array(length);
    result.set(input, length - input.length);
    return result;
  }

  /**
   * Converts a bigint to Uint8Array
   */
  private bigIntToUint8Array(value: bigint): Uint8Array {
    let hex = value.toString(16);
    if (hex.length % 2) hex = '0' + hex;
    const len = hex.length / 2;
    const result = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      result[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
    }
    return result;
  }

  /**
   * Hashes a series of Uint8Arrays and returns the result as a bigint
   */
  private async hashbigint(...arrays: Uint8Array[]): Promise<bigint> {
    const hash = await this.hashbyte(...arrays);
    return BigInt(
      '0x' +
        Array.from(hash)
          .map((b) => b.toString(16).padStart(2, '0'))
          .join('')
    );
  }

  /**
   * Concatenates and hashes multiple Uint8Arrays
   */
  private async hashbyte(...arrays: Uint8Array[]): Promise<Uint8Array> {
    const concatenated = this.concatUint8Arrays(...arrays);
    const hashBuffer = await crypto.subtle.digest(this.h, concatenated);
    return new Uint8Array(hashBuffer);
  }

  private concatUint8Arrays(...arrays: Uint8Array[]): Uint8Array {
    const totalLength = arrays.reduce((sum, arr) => sum + arr.length, 0);
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const arr of arrays) {
      result.set(arr, offset);
      offset += arr.length;
    }
    return result;
  }

  /**
   * Convert hex string to Uint8Array
   */
  private hexToUint8Array(hex: string): Uint8Array {
    if (hex.length % 2 !== 0) {
      hex = '0' + hex;
    }
    const result = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      result[i / 2] = parseInt(hex.slice(i, i + 2), 16);
    }
    return result;
  }

  private constantTimeEqual(a: Uint8Array, b: Uint8Array): boolean {
    if (a.length !== b.length) {
      return false;
    }

    let result = 0;
    for (let i = 0; i < a.length; i++) {
      result |= a[i] ^ b[i];
    }
    return result === 0;
  }
}
