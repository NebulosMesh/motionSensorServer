const HOST_URL = "http://localhost:8080/";

type ServiceName =
  | "getNodes"
  | "getNode"
  | "configureNode"
  | "configureAllNodes"
  | "requestHealth"
  | "getStatus"
  | "broadcastData"
  | "startServer"
  | "stopServer";

const endpoints: Record<ServiceName, string | ((mac: string) => string)> = {
  // node management
  getNodes: "/nodes",
  getNode: (mac: string) => `/nodes/${mac}`,
  configureNode: (mac: string) => `/nodes/${mac}/configure`,
  configureAllNodes: "/nodes/configure-all",
  // health and monitoring
  requestHealth: "/health/request",
  getStatus: "/status",
  // Data broadcasting
  broadcastData: "/broadcast",
  // server control
  startServer: "/server/start",
  stopServer: "/server/stop",
};

export default async function ApiService<T>(
  service: ServiceName,
  options?: RequestInit
): Promise<T> {
  const url = `${HOST_URL}${endpoints[service]}`;
  const response = await fetch(url, options);
  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }
  return response.json();
}

// Usage examples:
// callService('service_one');
// callService('service_two', { method: 'POST', body: JSON.stringify(data) });
