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

export interface CertificateOption {
  id: number;
  code: string;
  name: string;
  unlockType: string;
  threshold: number;
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

export interface DauPoint {
  date: string;
  activeUsers: number;
}

export interface RetentionStat {
  cohortUsers: number;
  returnedUsers: number;
  rate: number;
}

export interface ActivityData {
  dau: DauPoint[];
  retentionD1: RetentionStat | null;
  retentionD7: RetentionStat | null;
}

export interface RuleConfig {
  activationLbsThresholdMeters: number;
  activationDailyAttemptLimit: number;
  activationMaxSpeedMps: number;
  posterCampusBadge: string;
  checkinMinAmount: number;
  checkinMaxAmount: number;
  stealDailyTargets: number;
  stealDailyLimit: number;
  stealMinAmount: number;
  stealMaxAmount: number;
  giftDailyLimit: number;
  giftMinAmount: number;
  ironBonusThresholdMeters: number;
  rankingTopN: number;
  anomalyFeedDailyThreshold: number;
  anomalyStealDailyThreshold: number;
  anomalyListLimit: number;
  miniappUrl: string;
}

export interface FeedRankItem {
  rank: number;
  userId: number;
  nickname: string;
  total: number;
}

export interface CollegeRankItem {
  rank: number;
  collegeId: number;
  collegeName: string;
  total: number;
}

export async function getActivity(days = 14) {
  return requestClient.get<ActivityData>(
    sicauNiuApi("plugins/sicau-niu/settlement/activity"),
    { params: { days } },
  );
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

export async function getCertificateOptions() {
  const res = await requestClient.get<{ list: CertificateOption[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/certificates/options"),
  );
  return res.list ?? [];
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

export interface AnomalyAlert {
  userId: number;
  nickname: string;
  type: string;
  date: string;
  count: number;
  threshold: number;
}

export async function getRiskAnomalies() {
  const res = await requestClient.get<{ list: AnomalyAlert[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/risk/anomalies"),
  );
  return res.list ?? [];
}

export async function getRules() {
  return requestClient.get<RuleConfig>(
    sicauNiuApi("plugins/sicau-niu/settlement/rules"),
  );
}

export async function updateRules(data: RuleConfig) {
  return requestClient.put<RuleConfig>(
    sicauNiuApi("plugins/sicau-niu/settlement/rules"),
    data,
  );
}

export async function getFeedRanking() {
  const res = await requestClient.get<{ list: FeedRankItem[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/rankings/feed"),
  );
  return res.list ?? [];
}

export async function getFriendRanking() {
  const res = await requestClient.get<{ list: FeedRankItem[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/rankings/friend"),
  );
  return res.list ?? [];
}

export async function getCollegeRanking() {
  const res = await requestClient.get<{ list: CollegeRankItem[] }>(
    sicauNiuApi("plugins/sicau-niu/settlement/rankings/college"),
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
