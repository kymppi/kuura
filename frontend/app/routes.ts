import {
  type RouteConfig,
  index,
  prefix,
  route,
} from '@react-router/dev/routes';

export default [
  route('/', 'routes/index.tsx'),
  route('/home', 'routes/home.tsx'),
  ...prefix('login', [
    index('routes/login.tsx'),
    route(':serviceId', 'routes/login-serviceId.tsx'),
  ]),
] satisfies RouteConfig;
