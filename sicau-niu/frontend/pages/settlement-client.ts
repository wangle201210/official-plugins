import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface DashboardData {
  playerCount: number;
  activatedNiuCount: number;
  totalNiuCount: number;
  firstActivatorCount: number;
  feedingCount: number;
  feedTotalEffect: number;
  stealCount: number;
  giftCount: number;
  checkinCount: number;
  certificateGrantedCount: number;
}

export interface PlayerExportRow {
  userId: number;
  nickname: string;
  identityType: string;
  collegeName: string;
  grade: number;
  activationCount: number;
  feedTotalEffect: number;
}

export interface PlayerExportData {
  list: PlayerExportRow[];
  total: number;
  truncated: boolean;
}

export interface IssueResult {
  eligible: number;
  issued: number;
  skipped: number;
}

export interface DeviceClusterMember {
  userId: number;
  nickname: string;
}

export interface DeviceCluster {
  fingerprint: string;
  count: number;
  members: DeviceClusterMember[];
}

export interface ArchiveItem {
  id: number;
  title: string;
  snapshot: string;
  operatorId: number;
  archivedAt: number | null;
}

export async function getDashboard() {
  return requestClient.get<DashboardData>(
    sicauNiuApi("plugins/sicau-niu/settlement/dashboard"),
  );
}

export async function exportPlayers() {
  return requestClient.get<PlayerExportData>(
    sicauNiuApi("plugins/sicau-niu/settlement/export/players"),
  );
}

export async function issueCertificates(honorId: number) {
  return requestClient.post<IssueResult>(
    sicauNiuApi("plugins/sicau-niu/settlement/certificates/issue"),
    { honorId },
  );
}

export async function getRiskDeviceClusters() {
  const res = await requestClient.get<{ list: DeviceCluster[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/risk/device-clusters"),
  );
  return res.list ?? [];
}

export async function createArchive(title: string) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/settlement/archives"),
    { title },
  );
}

export async function listArchives() {
  const res = await requestClient.get<{ list: ArchiveItem[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/archives"),
  );
  return res.list ?? [];
}
