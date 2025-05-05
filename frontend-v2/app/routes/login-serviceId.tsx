import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router';
import type { Route } from './+types/home';

export function meta({}: Route.MetaArgs) {
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

  useEffect(() => {
    const checkAuthenticated = async () => {
      if (!serviceId) {
        navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
        return;
      }

      try {
        const response = await fetch('/v1/me', {
          credentials: 'include',
        });

        if (response.status === 200) {
          console.log(`Logging in to service: ${serviceId}`);
          //TODO: FETCH ACCESS TOKEN FOR THE SERVICE
          //TODO: REDIRECT TO SERVICE
        } else if (response.status === 401) {
          navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
        } else {
          console.error(`Unexpected status: ${response.status}`);
        }
      } catch (error) {
        console.error('Error checking auth:', error);
        navigate(`/login?return_to=${encodeURIComponent(location.pathname)}`);
      }
    };

    checkAuthenticated();
  }, [serviceId, navigate, location.pathname]);

  return (
    <div>
      <p>
        Checking authentication for service: <strong>{serviceId}</strong>
      </p>
    </div>
  );
}
