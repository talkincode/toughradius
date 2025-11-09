import simpleRestProvider from 'ra-data-simple-rest';
import { fetchUtils, DataProvider } from 'react-admin';

const API_BASE = '/api/v1';

const httpClient = (url: string, options: fetchUtils.Options = {}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: 'application/json' });
  }
  // 从 localStorage 获取 token
  const token = localStorage.getItem('token');
  if (token) {
    (options.headers as Headers).set('Authorization', `Bearer ${token}`);
  }
  return fetchUtils.fetchJson(url, options);
};

const resourcePathMap: Record<string, string> = {
  'radius/users': 'users',
  'radius/online': 'sessions',
  'radius/accounting': 'accounting',
  'radius/profiles': 'radius-profiles',
};

const resolveResource = (resource: string) =>
  resourcePathMap[resource] ?? resource;

const buildApiUrl = (resource: string, suffix = '') =>
  `${API_BASE}/${resolveResource(resource)}${suffix}`;

const extractData = <T = any>(json: any): T => json?.data ?? json;

const extractTotal = (json: any): number =>
  json?.meta?.total ??
  json?.total ??
  (Array.isArray(json?.data) ? json.data.length : Array.isArray(json) ? json.length : 0);

const baseDataProvider = simpleRestProvider(API_BASE, httpClient);

// 自定义 dataProvider 以适配后端 API 格式
export const dataProvider: DataProvider = {
  ...baseDataProvider,
  
  getList: async (resource, params) => {
    const { page = 1, perPage = 10 } = params.pagination || {};
    const { field = 'id', order = 'ASC' } = params.sort || {};
    const query = {
      sort: field,
      order,
      page,
      pageSize: perPage,
      ...params.filter,
    };

    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);

    return {
      data: extractData(json),
      total: extractTotal(json),
    };
  },

  getOne: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url);
    return { data: extractData(json) };
  },

  getMany: async (resource, params) => {
    const query = {
      filter: JSON.stringify({ id: params.ids }),
    };
    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);
    return { data: extractData(json) };
  },

  getManyReference: async (resource, params) => {
    const { page = 1, perPage = 10 } = params.pagination || {};
    const { field = 'id', order = 'ASC' } = params.sort || {};
    const query = {
      sort: field,
      order: order,
      page: page,
      pageSize: perPage,
      ...params.filter,
      [params.target]: params.id,
    };
    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);
    return {
      data: extractData(json),
      total: extractTotal(json),
    };
  },

  create: async (resource, params) => {
    const url = buildApiUrl(resource);
    const { json } = await httpClient(url, {
      method: 'POST',
      body: JSON.stringify(params.data),
    });
    const data = extractData(json);
    return {
      data: {
        ...params.data,
        ...(data ?? {}),
        id: data?.id ?? json?.id,
      } as any,
    };
  },

  update: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url, {
      method: 'PUT',
      body: JSON.stringify(params.data),
    });
    return { data: extractData(json) };
  },

  updateMany: async (resource, params) => {
    const responses = await Promise.all(
      params.ids.map(id =>
        httpClient(buildApiUrl(resource, `/${id}`), {
          method: 'PUT',
          body: JSON.stringify(params.data),
        })
      )
    );
    return {
      data: responses.map(({ json }, index) => {
        const data = extractData(json);
        return data?.id ?? json?.id ?? params.ids[index];
      }),
    };
  },

  delete: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url, {
      method: 'DELETE',
    });
    return { data: extractData(json) };
  },

  deleteMany: async (resource, params) => {
    const responses = await Promise.all(
      params.ids.map(id =>
        httpClient(buildApiUrl(resource, `/${id}`), {
          method: 'DELETE',
        })
      )
    );
    return {
      data: responses.map(({ json }, index) => {
        const data = extractData(json);
        return data?.id ?? json?.id ?? params.ids[index];
      }),
    };
  },
};
