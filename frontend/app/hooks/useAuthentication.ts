import { useContext } from 'react';
import { AuthContext } from '../contexts/AuthProvider';

export const useAuthentication = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuthentication must be used within an AuthProvider');
  }
  return context;
};
