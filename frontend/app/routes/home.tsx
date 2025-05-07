import { Button, Loading, Stack } from '@carbon/react';
import { useNavigate } from 'react-router';
import { useAuthentication } from '../hooks/useAuthentication';

export function meta() {
  return [
    { title: 'Kuura' },
    {
      name: 'description',
      content: 'A simpleish user and authentication provider.',
    },
  ];
}

export default function Home() {
  const navigate = useNavigate();
  const { loading, authenticated, user, client } = useAuthentication();

  if (loading) {
    return (
      <div
        style={{
          display: 'grid',
          width: '100%',
          height: '100%',
          placeItems: 'center',
        }}
      >
        <Loading />
      </div>
    );
  }

  if (!user || !authenticated) {
    navigate('/login');
    return <h1>Unauthorized (or no user found).</h1>;
  }

  return (
    <Stack>
      <h1>Welcome, {user?.username}!</h1>
      <Button kind="secondary" onClick={() => client.refreshAccessToken()}>
        Refresh tokens
      </Button>
      <Button kind="danger" onClick={() => client.logout()}>
        Log out
      </Button>
    </Stack>
  );
}
