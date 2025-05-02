import { Button, Form, PasswordInput, Stack, TextInput } from '@carbon/react';

export default function LoginPage() {
  return (
    <div
      style={{
        display: 'grid',
        placeItems: 'center',
        height: '100%',
        padding: '1rem',
        backgroundImage: 'url("/login.jpg")',
        backgroundSize: 'cover',
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center',
      }}
    >
      <LoginForm />
    </div>
  );
}

function LoginForm() {
  return (
    <Form
      style={{
        width: '100%',
        maxWidth: '32rem',
        padding: '1rem',
        backgroundColor: 'white',
      }}
    >
      <Stack gap={7}>
        <Stack gap="0.25rem">
          <h1>Log in to [service]</h1>
          <p>
            Don't have an account? Contact the{' '}
            <a href="/">[instance administrator]</a>.
          </p>
        </Stack>
        <Stack gap="1rem">
          <TextInput
            id="username"
            labelText="Username"
            placeholder="Enter your username"
            required
          />

          <PasswordInput
            id="password"
            labelText="Password"
            placeholder="Enter your password"
            required
          />
        </Stack>

        <Button
          style={{
            justifySelf: 'flex-end',
          }}
          type="submit"
        >
          Log in
        </Button>
      </Stack>
    </Form>
  );
}
