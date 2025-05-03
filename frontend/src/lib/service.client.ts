export interface ServiceInfo {
  name: string;
  contact: {
    name: string;
    link: string;
  };
}

export const getServiceInfo = async (
  serviceId: string
): Promise<ServiceInfo> => {
  const data = await fetch(`/v1/${serviceId}`);
  if (!data.ok) {
    throw new Error(data.statusText);
  }

  const response = (await data.json()) as {
    name: string;
    contact: string;
    contact_email: string;
  };

  return {
    name: response.name,
    contact: {
      name: response.contact,
      link: `mailto:${response.contact_email}`,
    },
  };
};
