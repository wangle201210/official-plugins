import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface CardItem {
  id: number;
  niuId: number;
  niuCode: string;
  niuName: string;
  category: string;
  title: string;
  content: string;
  imagePath: string;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface CardListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  category?: string;
}

export interface CardSaveInput {
  niuId: number;
  category: string;
  title: string;
  content: string;
  imagePath: string;
}

export async function listCards(params?: CardListParams) {
  const res = await requestClient.get<{
    list: CardItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/cards"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function getCard(id: number) {
  return requestClient.get<CardItem>(
    sicauNiuApi(`plugins/sicau-niu/admin/cards/${id}`),
  );
}

export async function createCard(values: CardSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/cards"),
    values,
  );
}

export async function updateCard(id: number, values: CardSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/cards/${id}`),
    values,
  );
}

export async function deleteCard(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/cards/${id}`),
  );
}
