import { useFetcher } from "react-router";
import type { Route } from "../+types/root";
import ApiService, { dev_ApiService } from "../services/apiService";
import type { IApiResponse } from "~/interfaces/IApiService";
import { useState } from "react";
import { formatTime } from "~/services/formatDateTime";

export async function loader({ request }: Route.LoaderArgs) {
  // Get server status from API
     const status = await ApiService<{ status: string }>("getStatus");
  //const response = (await dev_ApiService("getStatus")) as IApiResponse;
  console.log("SERVER STATUS: ", status);

  return status;
}

export default function Server({ loaderData }: { loaderData?: IApiResponse }) {
  const fetcher = useFetcher<{ status: string }>();
  const [serverData, setServerData] = useState(
    loaderData?.data ?? { running: false }
  );

  // Show loading state if fetcher is submitting
  const isSubmitting = fetcher.state === "submitting";

  return (
    <div className="max-w-md mx-auto mt-10 p-8 bg-emerald-700 rounded-lg shadow-md text-center">
      <h1 className="text-3xl font-bold mb-4">Server Statuses</h1>
      <p className="text-lg mb-6">
        Running: <span>{serverData.running.toString()}</span>
      </p>
      <p className="text-lg mb-6">
        Total Nodes: <span>{serverData.totalNodes}</span>
      </p>
      <p className="text-lg mb-6">
        Online Nodes: <span>{serverData.onlineNodes}</span>
      </p>
      <p className="text-lg mb-6">
        Last Checked: <span>{formatTime(serverData.timestamp)}</span>
      </p>
      <fetcher.Form method="post" className="flex gap-4 justify-center mb-4">
        <button
          type="submit"
          name="action"
          value="start"
          disabled={isSubmitting}
          className={`px-4 py-2 rounded bg-blue-600 text-white font-medium transition hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed`}>
          Start Server
        </button>
        <button
          type="submit"
          name="action"
          value="stop"
          disabled={isSubmitting}
          className={`px-4 py-2 rounded bg-red-600 text-white font-medium transition hover:bg-red-700 disabled:opacity-60 disabled:cursor-not-allowed`}>
          Stop Server
        </button>
      </fetcher.Form>
      {isSubmitting && <p className="mt-2 animate-pulse">Processing...</p>}
    </div>
  );
}

// Add an action to handle start/stop requests
export async function action({ request }: Route.ActionArgs) {
  const formData = await request.formData();
  const actionType = formData.get("action");

  if (actionType === "start") {
    await ApiService("startServer", { method: "POST" });
  } else if (actionType === "stop") {
    await ApiService("stopServer", { method: "POST" });
  }

  // Return updated status after action
  const status = await ApiService<{ status: string }>("getStatus");
  return { status };
}
