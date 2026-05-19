import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0182 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-182 租户删除生命周期前置条件', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-182a: active tenant delete passes lifecycle precondition before soft delete', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0182();
  });
});
