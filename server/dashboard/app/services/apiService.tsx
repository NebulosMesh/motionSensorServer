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

export default async function ApiService<ApiResponse>(
  service: ServiceName,
  options?: RequestInit
): Promise<ApiResponse> {
  const url = `${HOST_URL}${endpoints[service]}`;
  const response: Response = await fetch(url, options);
  if (!response.ok) {
    throw new Error(
      `API error: ${response.status ?? "Unknown Error occurred"}`
    );
  }
  return response.json();
}

// Use this one during deving with dummy data

export async function dev_ApiService<ApiResponse>(
  service: ServiceName,
  options?: RequestInit
): Promise<ApiResponse> {
  const dev_endpoints = {
    // node management
    getNodes: {
      success: true,
      data: [
        {
          mac: "AA:BB:CC:DD:EE:FF",
          name: "Node 1",
          online: true,
          lastSeen: 1693400000,
        },
        {
          mac: "11:22:33:44:55:66",
          name: "Node 2",
          online: false,
          lastSeen: 1693399000,
        },
        {
          mac: "AA:BB:CC:44:55:66",
          name: "Node 3",
          online: true,
          lastSeen: 1693400000,
        },
        {
          mac: "DD:EE:FF:11:22:33",
          name: "Node 4",
          online: false,
          lastSeen: 1693399000,
        },
      ],
    },
    getNode: {
      success: true,
      data: {
        mac: "AA:BB:CC:DD:EE:FF",
        name: "Node 1",
        online: true,
        lastSeen: 1693400000,
        adapterType: 1,
      },
    },
    configureNode: {
      success: true,
      message: "Node AA:BB:CC:DD:EE:FF configured to adapter type WiFi",
    },
    configureAllNodes: {
      success: true,
      message: "All nodes configured to adapter type Bluetooth",
    },
    // health and monitoring
    requestHealth: {
      success: true,
      message: "Health reports requested",
    },
    getStatus: {
      success: true,
      data: {
        running: true,
        totalNodes: 5,
        onlineNodes: 3,
        timestamp: 1693400100,
      },
    },
    // Data broadcasting
    broadcastData: {
      success: true,
      message: "Data broadcasted to all nodes (type: WiFi, length: 128)",
    },
    // server control
    startServer: {
      success: true,
      message: "Mesh server started",
    },
    stopServer: {
      success: true,
      message: "Mesh server stopped",
    },
    errorResponse: {
      success: false,
      error: "This is an error message. Ohh no..",
    },
  };

  return dev_endpoints[service] as ApiResponse;
}

// Usage examples:
// callService('service_one');
// callService('service_two', { method: 'POST', body: JSON.stringify(data) });
