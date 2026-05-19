import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';

/** 查询表单schema */
export function buildQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('plugin.linapro-monitor-online.page.fields.userAccount'),
    },
    {
      component: 'Input',
      fieldName: 'ip',
      label: $t('plugin.linapro-monitor-online.page.fields.ipAddress'),
    },
  ];
}

/** 表格列配置 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    {
      title: $t('plugin.linapro-monitor-online.page.fields.loginAccount'),
      field: 'username',
    },
    {
      title: $t('plugin.linapro-monitor-online.page.fields.departmentName'),
      field: 'deptName',
    },
    {
      title: $t('plugin.linapro-monitor-online.page.fields.ipAddress'),
      field: 'ip',
    },
    {
      title: $t('plugin.linapro-monitor-online.page.fields.browser'),
      field: 'browser',
    },
    {
      title: $t('plugin.linapro-monitor-online.page.fields.os'),
      field: 'os',
    },
    {
      title: $t('plugin.linapro-monitor-online.page.fields.loginTime'),
      field: 'loginTime',
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('plugin.linapro-monitor-online.page.fields.actions'),
      resizable: false,
      width: 120,
    },
  ];
}
