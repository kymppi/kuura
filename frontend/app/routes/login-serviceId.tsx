import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { useAuthentication } from '../hooks/useAuthentication';
import type { Route } from './+types/home';

export function meta() {
  return [
    { title: 'Kuura' },
    {
      name: 'description',
      content: 'Log in to Kuura.',
    },
  ];
}

export default function ServiceLogin({ params }: Route.ComponentProps) {
  const { serviceId } = params;
  const navigate = useNavigate();
  const location = useLocation();
  const { authenticated, loading, client } = useAuthentication();

  useEffect(() => {
    if (loading) return;

    const performLogin = async () => {
      if (!serviceId) {
        navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
        return;
      }

      if (authenticated) {
        console.log(`Logging in to service: ${serviceId}`);
        try {
          const redirectUrl = await client.loginToService(serviceId);

          if (!redirectUrl) {
            console.error('No redirect URL returned from service login');
          } else {
            window.location.href = redirectUrl;
          }
        } catch (error) {
          console.error('Failed to login to service:', error);
          navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
        }
      } else {
        navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
      }
    };

    performLogin();
  }, [serviceId, authenticated, loading, navigate, location.pathname]);

  return (
    <div>
      <p>
        Logging into: <strong>{serviceId}</strong>
      </p>
    </div>
  );
}
