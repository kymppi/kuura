export interface ServiceInfo {
  id: string;
  name: string;
  contact: {
    name: string;
    link: string;
  };
  redirects: {
    login: string;
  };
}

export const getServiceInfo = async (
  serviceId?: string
): Promise<ServiceInfo> => {
  const id = serviceId ? serviceId : 'kuura';

  const data = await fetch(`/v1/service/${id}`);
  if (!data.ok) {
    throw new Error(data.statusText);
  }

  const response = (await data.json()) as {
    id: string;
    name: string;
    contact: string;
    contact_email: string;
    login_redirect: string;
  };

  return {
    id: response.id,
    name: response.name,
    contact: {
      name: response.contact,
      link: `mailto:${response.contact_email}`,
    },
    redirects: {
      login: response.login_redirect,
    },
  };
};
