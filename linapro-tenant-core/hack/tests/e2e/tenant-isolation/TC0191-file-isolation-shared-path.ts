import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0191 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-191 文件跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-191a: file storage paths include tenant buckets and platform shared paths', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0191();
  });
});
