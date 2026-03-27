import { createRouter, createWebHashHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { translate } from '../i18n'
import ClientConfigure from '../views/ClientConfigure.vue'
import ClientSettings from '../views/ClientSettings.vue'
import ProxyEdit from '../views/ProxyEdit.vue'
import VisitorEdit from '../views/VisitorEdit.vue'
import PairingCreate from '../views/PairingCreate.vue'
import PairingImport from '../views/PairingImport.vue'
import { useProxyStore } from '../stores/proxy'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/settings',
    },
    {
      path: '/settings',
      name: 'ClientSettings',
      component: ClientSettings,
    },
    {
      path: '/proxies',
      name: 'ProxyList',
      component: () => import('../views/ProxyList.vue'),
    },
    {
      path: '/proxies/detail/:name',
      name: 'ProxyDetail',
      component: () => import('../views/ProxyDetail.vue'),
    },
    {
      path: '/proxies/create',
      name: 'ProxyCreate',
      component: ProxyEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/proxies/:name/edit',
      name: 'ProxyEdit',
      component: ProxyEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/visitors',
      name: 'VisitorList',
      component: () => import('../views/VisitorList.vue'),
    },
    {
      path: '/visitors/detail/:name',
      name: 'VisitorDetail',
      component: () => import('../views/VisitorDetail.vue'),
    },
    {
      path: '/visitors/create',
      name: 'VisitorCreate',
      component: VisitorEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/visitors/:name/edit',
      name: 'VisitorEdit',
      component: VisitorEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/config',
      name: 'ClientConfigure',
      component: ClientConfigure,
    },
    {
      path: '/pairing/create',
      name: 'PairingCreate',
      component: PairingCreate,
      meta: { requiresStore: true },
    },
    {
      path: '/pairing/import',
      name: 'PairingImport',
      component: PairingImport,
    },
  ],
})

router.beforeEach(async (to) => {
  if (!to.matched.some((record) => record.meta.requiresStore)) {
    return true
  }

  const proxyStore = useProxyStore()
  const enabled = await proxyStore.checkStoreEnabled()
  if (enabled) {
    return true
  }

  ElMessage.warning(
    translate('router.storeDisabled'),
  )
  return { name: 'ClientSettings' }
})

export default router
