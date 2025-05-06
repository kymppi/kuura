import {
  type RouteConfig,
  index,
  prefix,
  route,
} from '@react-router/dev/routes';

export default [
  index('routes/home.tsx'),
  ...prefix('login', [
    index('routes/login.tsx'),
    route(':serviceId', 'routes/login-serviceId.tsx'),
  ]),
] satisfies RouteConfig;
