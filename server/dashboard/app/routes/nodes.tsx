import ApiService from "~/services/apiService";
import type { Route } from "../+types/root";

export async function loader({ request }: Route.LoaderArgs) {
  //   const nodes = ApiService("getNodes");
  const nodes = [
    { id: "1", status: "active" },
    { id: "2", status: "inactive" },
    { id: "3", status: "active" },
    { id: "4", status: "inactive" },
    { id: "5", status: "active" },
    { id: "6", status: "inactive" },
  ];
  return nodes;
}

type Node = { id: string; status: string };

export default function Nodes({ loaderData }: Route.ComponentProps) {
  const nodes = loaderData as unknown as Node[];
  return (
    <div className="p-4 justify-center">
      <h1 className="text-center">Nodes</h1>
      <br />
      <div className="nodes-container grid grid-cols-3 gap-2">
        {nodes?.map((node, index) => (
          <div className="node-card h-[100px] rounded-2xl bg-gray-500 p-2">
            <p>Node {node.id}</p>
            <p>status: {node.status}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
