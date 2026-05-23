import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { LanguageEnum } from '@/enums/appEnum'
import { router } from '@/router'
import { useSettingStore } from './setting'
import { useWorktabStore } from './worktab'
import { AppRouteRecord } from '@/types/router'
import { setPageTitle } from '@/utils/router'
import { resetRouterState } from '@/router/guards/beforeEach'
import { useMenuStore } from './menu'
import { StorageConfig } from '@/utils/storage/storage-config'
import { fetchLogout } from '@/api/auth'

export const useUserStore = defineStore(
  'userStore',
  () => {
    const language = ref(LanguageEnum.ZH)
    const isLogin = ref(false)
    const isLock = ref(false)
    const lockPassword = ref('')
    const info = ref<Partial<Api.Auth.UserInfo>>({})
    const searchHistory = ref<AppRouteRecord[]>([])
    const accessToken = ref('')

    const getUserInfo = computed(() => info.value)
    const getSettingState = computed(() => useSettingStore().$state)
    const getWorktabState = computed(() => useWorktabStore().$state)

    const setUserInfo = (newInfo: Api.Auth.UserInfo | Api.Auth.MeResponse) => {
      info.value = 'session_id' in newInfo ? normalizeUserInfo(newInfo) : newInfo
    }

    const setLoginStatus = (status: boolean) => {
      isLogin.value = status
    }

    const setLanguage = (lang: LanguageEnum) => {
      setPageTitle(router.currentRoute.value)
      language.value = lang
    }

    const setSearchHistory = (list: AppRouteRecord[]) => {
      searchHistory.value = list
    }

    const setLockStatus = (status: boolean) => {
      isLock.value = status
    }

    const setLockPassword = (password: string) => {
      lockPassword.value = password
    }

    const setToken = (newAccessToken: string) => {
      accessToken.value = newAccessToken
    }

    const logOut = async (options?: { redirect?: boolean; revoke?: boolean }) => {
      const shouldRedirect = options?.redirect !== false
      const shouldRevoke = options?.revoke !== false

      if (shouldRevoke && (accessToken.value || isLogin.value)) {
        try {
          await fetchLogout()
        } catch (error) {
          console.warn('[Auth] logout request failed', error)
        }
      }

      const currentUserId = info.value.userId
      if (currentUserId) {
        localStorage.setItem(StorageConfig.LAST_USER_ID_KEY, String(currentUserId))
      }

      info.value = {}
      isLogin.value = false
      isLock.value = false
      lockPassword.value = ''
      accessToken.value = ''

      sessionStorage.removeItem('iframeRoutes')
      useMenuStore().setHomePath('')
      resetRouterState(500)

      if (shouldRedirect) {
        const currentRoute = router.currentRoute.value
        const redirect = currentRoute.path !== '/login' ? currentRoute.fullPath : undefined
        router.push({
          name: 'Login',
          query: redirect ? { redirect } : undefined
        })
      }
    }

    const checkAndClearWorktabs = () => {
      const lastUserId = localStorage.getItem(StorageConfig.LAST_USER_ID_KEY)
      const currentUserId = info.value.userId

      if (!currentUserId) return
      if (!lastUserId) return

      if (String(currentUserId) !== lastUserId) {
        const worktabStore = useWorktabStore()
        worktabStore.opened = []
        worktabStore.keepAliveExclude = []
      }

      localStorage.removeItem(StorageConfig.LAST_USER_ID_KEY)
    }

    return {
      language,
      isLogin,
      isLock,
      lockPassword,
      info,
      searchHistory,
      accessToken,
      getUserInfo,
      getSettingState,
      getWorktabState,
      setUserInfo,
      setLoginStatus,
      setLanguage,
      setSearchHistory,
      setLockStatus,
      setLockPassword,
      setToken,
      logOut,
      checkAndClearWorktabs
    }
  },
  {
    persist: {
      key: 'user',
      storage: localStorage
    }
  }
)

function normalizeUserInfo(payload: Api.Auth.MeResponse): Api.Auth.UserInfo {
  const permissions = [...(payload.user.permissions ?? [])]
  const publicID = payload.user.id
  return {
    id: publicID,
    userId: publicID,
    userName: publicID,
    email: '',
    roles: [...payload.user.roles],
    permissions,
    buttons: permissions,
    sessionId: payload.session_id
  }
}
