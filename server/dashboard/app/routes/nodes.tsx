import ApiService, { dev_ApiService } from "~/services/apiService";
import type { IApiResponse } from "~/interfaces/IApiService";
import type { Route } from "../+types/root";
import type { INode, INodes } from "~/interfaces/INodes";
import NodeCard from "~/components/NodeCard/nodeCard";

export async function loader({}: Route.LoaderArgs) {
  // const response = (await ApiService("getNodes")) as ApiResponse;
  const response = (await dev_ApiService("getNodes")) as IApiResponse;
  console.log("NODES: ", response);

  if (!response.success) {
    throw new Response(response.error, {
      status: 500,
    });
  }
  if (response.data === undefined) {
    console.log("No nodes found");
    throw new Response(response.error, {
      status: 404,
    });
  }

  return (response.data as INodes) || [];
}

export default function Nodes({ loaderData }: Route.ComponentProps) {
  const nodes = loaderData as INodes | undefined;
  return (
    <div className="p-6 justify-center">
      <h1 className="text-center">Nodes</h1>
      <br />
      <div className="nodes-container w-[80%] grid grid-cols-3 gap-4 justify-center m-auto">
        {nodes?.map((node: INode, index: any) => (
          <NodeCard key={index} nodeData={node} />
        ))}
      </div>
    </div>
  );
}
