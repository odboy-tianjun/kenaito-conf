import request from './request'

export function getClientListByConfig(configId) {
  return request({
    url: '/admin/client/listByConfig',
    method: 'get',
    params: { configId }
  })
}
