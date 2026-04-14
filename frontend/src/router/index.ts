import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { public: true } },

  { path: '/dashboard',     name: 'dashboard',     component: () => import('@/views/Dashboard.vue') },
  { path: '/regulatory',    name: 'regulatory',    component: () => import('@/views/Regulatory.vue') },
  { path: '/stakeholders',  name: 'stakeholders',  component: () => import('@/views/Stakeholders.vue') },
  { path: '/membership',    name: 'membership',    component: () => import('@/views/Membership.vue') },
  { path: '/consultations', name: 'consultations', component: () => import('@/views/Consultations.vue') },
  { path: '/calendar',      name: 'calendar',      component: () => import('@/views/Calendar.vue') },
  { path: '/activities',    name: 'activities',    component: () => import('@/views/Activities.vue') },
  { path: '/templates',     name: 'templates',     component: () => import('@/views/Templates.vue') },
  { path: '/people',        name: 'people',        component: () => import('@/views/People.vue') },
  { path: '/engagement',    name: 'engagement',    component: () => import('@/views/Engagement.vue') },
  { path: '/members',       name: 'members',       component: () => import('@/views/Members.vue') },
  { path: '/members/:id',   name: 'member-detail', component: () => import('@/views/MemberDetail.vue'), props: true },
  { path: '/audit',         name: 'audit',         component: () => import('@/views/AuditLog.vue'), meta: { role: 'admin' } },

  { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/views/NotFound.vue') },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() { return { top: 0 }; },
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  if (!to.meta.public && !auth.token) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  const requiredRole = to.meta.role as string | undefined;
  if (requiredRole && !auth.hasRole(requiredRole)) {
    return { name: 'dashboard' };
  }
  return true;
});

export default router;
