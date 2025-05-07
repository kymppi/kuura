import { createContext, useEffect, useState } from 'react';
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
}>({
  user: null,
  client: {} as SRPAuthClient,
  authenticated: false,
  loading: true,
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

  useEffect(() => {
    client
      .getUser()
      .then((userData) => {
        setUser(userData);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching user:', error);
        setLoading(false);
      });
  }, [client]);

  const authenticated = !!user;

  return (
    <AuthContext.Provider value={{ user, client, authenticated, loading }}>
      {children}
    </AuthContext.Provider>
  );
};
