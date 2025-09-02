export interface INode {
  mac: string;
  online: boolean;
  name: string;
  lastSeen: string;
}

export type INodes = INode[];

export interface INodeCardProps {
  key: React.Key;
  nodeData: INode;
}

// Example Data:
// [
//   {
//     mac: "AA:BB:CC:DD:EE:FF",
//     name: "Node 1",
//     online: true,
//     lastSeen: 1693400000,
//   },
//   {
//     mac: "11:22:33:44:55:66",
//     name: "Node 2",
//     online: false,
//     lastSeen: 1693399000,
//   },
// ];
