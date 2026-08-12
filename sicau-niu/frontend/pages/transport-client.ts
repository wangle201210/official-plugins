import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function transportApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface TransportTeamItem {
  id: string;
  name: string;
  status: "effective" | "invalid";
  creatorId: string;
  creatorName: string;
  memberCount: number;
  totalContributionMeters: number;
  lastActiveAt: number;
  invalidatedAt?: number;
  invalidReason: string;
  createdAt: number;
}

export interface TransportReportItem {
  id: string;
  teamId: string;
  teamName: string;
  memberId: string;
  userId: string;
  userName: string;
  activityDate: string;
  startLat?: number;
  startLng?: number;
  endLat: number;
  endLng: number;
  sampledAt: number;
  acceptedAt: number;
  contributionMeters: number;
  userTotalMeters: number;
  teamTotalMeters: number;
}

export interface TransportStats {
  effectiveTeamCount: number;
  invalidTeamCount: number;
  activeMemberCount: number;
  reportCount: number;
  contributionMeters: number;
}

interface TransportErrorEnvelope {
  errorCode?: string;
  message?: string;
}

interface PagedParams {
  pageNum?: number;
  pageSize?: number;
}

export async function listTransportTeams(
  params: PagedParams & { name?: string; status?: string },
) {
  const res = await requestClient.get<{
    list: TransportTeamItem[];
    total: number;
  }>(transportApi("plugins/sicau-niu/admin/transport-teams"), { params });
  return { items: res.list ?? [], total: res.total ?? 0 };
}

export function updateTransportTeamName(id: string, name: string) {
  return requestClient.put(
    transportApi(`plugins/sicau-niu/admin/transport-teams/${id}/name`),
    { name },
    { silentErrorMessage: true },
  );
}

export function transportRenameErrorMessage(error: unknown) {
  const envelope = (error as { response?: { data?: TransportErrorEnvelope } })
    ?.response?.data;
  if (
    envelope?.errorCode ===
    "PLUGIN_SICAU_NIU_TRANSPORT_TEAM_NAME_TAKEN"
  ) {
    return "这个有效团名已经被使用";
  }
  return envelope?.message || "团名称更新失败，请稍后重试";
}

export function getTransportStats(activityDate?: string) {
  return requestClient.get<TransportStats>(
    transportApi("plugins/sicau-niu/admin/transport-stats"),
    { params: { activityDate } },
  );
}

export async function listTransportReports(
  params: PagedParams & {
    activityDate?: string;
    teamId?: number;
    userId?: number;
  },
) {
  const res = await requestClient.get<{
    list: TransportReportItem[];
    total: number;
  }>(transportApi("plugins/sicau-niu/admin/transport-reports"), { params });
  return { items: res.list ?? [], total: res.total ?? 0 };
}
