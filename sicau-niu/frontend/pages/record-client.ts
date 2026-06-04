import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function recordApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

interface PagedParams {
  pageNum?: number;
  pageSize?: number;
}

async function listRecords<T>(pathName: string, params: Record<string, any>) {
  const res = await requestClient.get<{ list: T[]; total: number }>(
    recordApi(pathName),
    { params },
  );
  return { items: res.list ?? [], total: res.total ?? 0 };
}

export interface FeedingRecord {
  id: number;
  userId: number;
  nickname: string;
  niuId: number;
  niuName: string;
  niuCode: string;
  baseAmount: number;
  coefficientBasis: number;
  effectAmount: number;
  isIronBonus: number;
  fedAt: number | null;
  createdAt: number | null;
}

export interface StealRecord {
  id: number;
  actorUserId: number;
  actorNickname: string;
  targetUserId: number;
  targetNickname: string;
  amount: number;
  stealDate: string;
  createdAt: number | null;
}

export interface GiftRecord {
  id: number;
  fromUserId: number;
  fromNickname: string;
  toUserId: number;
  toNickname: string;
  amount: number;
  giftDate: string;
  createdAt: number | null;
}

export interface CheckinRecord {
  id: number;
  userId: number;
  nickname: string;
  checkinDate: string;
  amount: number;
  createdAt: number | null;
}

export interface ActivationRecord {
  id: number;
  userId: number;
  nickname: string;
  niuId: number;
  niuName: string;
  niuCode: string;
  activityDate: string;
  isFirst: number;
  orderNo: number;
  activatedAt: number | null;
  createdAt: number | null;
}

export interface GrassTxnRecord {
  id: number;
  userId: number;
  nickname: string;
  delta: number;
  txnType: string;
  refId: number;
  createdAt: number | null;
}

export function listFeedings(
  params: PagedParams & { niuId?: number; userId?: number },
) {
  return listRecords<FeedingRecord>(
    "plugins/sicau-niu/admin/records/feedings",
    params,
  );
}

export function listSteals(
  params: PagedParams & { actorUserId?: number; targetUserId?: number },
) {
  return listRecords<StealRecord>(
    "plugins/sicau-niu/admin/records/steals",
    params,
  );
}

export function listGifts(
  params: PagedParams & { fromUserId?: number; toUserId?: number },
) {
  return listRecords<GiftRecord>(
    "plugins/sicau-niu/admin/records/gifts",
    params,
  );
}

export function listCheckins(params: PagedParams & { userId?: number }) {
  return listRecords<CheckinRecord>(
    "plugins/sicau-niu/admin/records/checkins",
    params,
  );
}

export function listActivations(
  params: PagedParams & { userId?: number; niuId?: number },
) {
  return listRecords<ActivationRecord>(
    "plugins/sicau-niu/admin/records/activations",
    params,
  );
}

export function listGrassTxns(params: PagedParams & { userId?: number }) {
  return listRecords<GrassTxnRecord>(
    "plugins/sicau-niu/admin/records/grass-txns",
    params,
  );
}
