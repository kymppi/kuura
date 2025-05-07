import { createContext, useEffect, useMemo, useRef, useState } from 'react';
import { SRPAuthClient } from '../lib/auth.client';

export interface User {
  id: string;
  username: string;
  last_login_at: Date;
}

export const AuthContext = createContext<{
  user: User | null;
  client: SRPAuthClient;
  authenticated: boolean;
  loading: boolean;
  refreshUser: () => Promise<boolean>;
}>({
  user: null,
  client: {} as SRPAuthClient,
  authenticated: false,
  loading: true,
  refreshUser: () => {
    return Promise.resolve(false);
  },
});

export const AuthProvider = ({
  children,
  client,
}: {
  readonly children: React.ReactNode;
  readonly client: SRPAuthClient;
}) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const userFetched = useRef(false);

  useEffect(() => {
    if (client) {
      setLoading(true);
      userFetched.current = false;
    }
  }, [client]);

  useEffect(() => {
    if (loading && !userFetched.current && client) {
      userFetched.current = true;

      client
        .getUser()
        .then((userData) => {
          setUser(userData);
          setLoading(false);
        })
        .catch((error) => {
          console.error('Error fetching user:', error);
          setUser(null);
          setLoading(false);
        });
    }
  }, [client, loading]);

  const refreshUser = async () => {
    try {
      setLoading(true);
      const userData = await client.getUser();
      setUser(userData);
      return true;
    } catch (error) {
      console.error('Error refreshing user:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const authenticated = !!user;

  const value = useMemo(() => ({
    user,
    client,
    authenticated,
    loading,
    refreshUser,
  }), [user, client, authenticated, loading, refreshUser]);

  return (
    <AuthContext.Provider
      value={value}
    >
      {children}
    </AuthContext.Provider>
  );
};
