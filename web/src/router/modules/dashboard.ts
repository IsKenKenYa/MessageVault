import { AppRouteRecord } from '@/types/router'

export const dashboardRoutes: AppRouteRecord = {
  name: 'Dashboard',
  path: '/dashboard',
  component: '/index/index',
  meta: {
    title: 'menus.dashboard.title',
    icon: 'ri:pie-chart-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'console',
      name: 'Console',
      component: '/dashboard/console',
      meta: {
        title: 'menus.dashboard.overview',
        keepAlive: false,
        fixedTab: true
      }
    },
    {
      path: 'timeline',
      name: 'Timeline',
      component: '/dashboard/timeline',
      meta: {
        title: 'menus.dashboard.timeline',
        keepAlive: true
      }
    },
    {
      path: 'search',
      name: 'SearchRecords',
      component: '/dashboard/search',
      meta: {
        title: 'menus.dashboard.search',
        keepAlive: true
      }
    },
    {
      path: 'contacts',
      name: 'Contacts',
      component: '/dashboard/contacts',
      meta: {
        title: 'menus.dashboard.contacts',
        keepAlive: true
      }
    },
    {
      path: 'imports',
      name: 'Imports',
      component: '/dashboard/imports',
      meta: {
        title: 'menus.dashboard.imports',
        keepAlive: true
      }
    }
  ]
}
