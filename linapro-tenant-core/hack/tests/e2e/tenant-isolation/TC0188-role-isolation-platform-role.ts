import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0188 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-188 角色跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-188a: tenant roles and platform roles persist in disjoint tenant buckets', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0188();
  });
});
