import request from '@/utils/request';

export function getDnsRecords() {
  return request({
    url: '/api/dns/getall',
    method: 'get'
  });
}

export function createDnsZone(data: unknown) {
  return request({
    url: '/api/dns/zone/add',
    method: 'post',
    data
  });
}

export function updateDnsZone(data: unknown) {
  return request({
    url: '/api/dns/zone/update',
    method: 'post',
    data
  });
}

export function batchDeleteDnsZoneByIds(data: unknown) {
  return request({
    url: '/api/dns/zone/delete',
    method: 'post',
    data
  });
}

export function createDnsRecord(data: unknown) {
  return request({
    url: '/api/dns/record/add',
    method: 'post',
    data
  });
}

export function updateDnsRecord(data: unknown) {
  return request({
    url: '/api/dns/record/update',
    method: 'post',
    data
  });
}

export function batchDeleteDnsRecordByIds(data: unknown) {
  return request({
    url: '/api/dns/record/delete',
    method: 'post',
    data
  });
}
