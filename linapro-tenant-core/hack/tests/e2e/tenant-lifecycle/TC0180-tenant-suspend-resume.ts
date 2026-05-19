import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0180 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-180 租户暂停恢复', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-180a: suspended tenant blocks login and resumed tenant allows it again', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0180();
  });
});
