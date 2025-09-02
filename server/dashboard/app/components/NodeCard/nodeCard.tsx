import type { INodeCardProps } from "~/interfaces/INodes";
import { formatTime } from "~/services/formatDateTime";

export default function NodeCard({ key, nodeData }: INodeCardProps) {
  return (
    <div key={key} className="node-card h-max rounded-2xl bg-emerald-700 p-3">
      <p className="text-center text-2xl">{nodeData.name}</p>
      <p className="text-xs text-right">mac: {nodeData.mac}</p>
      <p>Online?: {nodeData.online ? "yes" : "no"}</p>
      <p>last Seen: {formatTime(nodeData.lastSeen)}</p>
    </div>
  );
}
