import {
  AnchorHTMLAttributes,
  Children,
  ReactNode,
  createContext,
  isValidElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState
} from 'react'

type RouterContextValue = {
  pathname: string
  navigate: (to: string, options?: { replace?: boolean }) => void
}

const RouterContext = createContext<RouterContextValue | null>(null)

const normalizePath = (value: string) => {
  if (!value) {
    return '/'
  }

  const [path] = value.split('?')
  const trimmed = path.replace(/\/+$/, '')
  if (trimmed === '' || trimmed === '/') {
    return '/'
  }

  return trimmed.startsWith('/') ? trimmed : `/${trimmed}`
}

const getInitialPath = () => {
  if (typeof window === 'undefined') {
    return '/'
  }

  return normalizePath(window.location.pathname)
}

type BrowserRouterProps = {
  children: ReactNode
}

export function BrowserRouter({ children }: BrowserRouterProps) {
  const [pathname, setPathname] = useState<string>(() => getInitialPath())

  const navigate = useCallback<RouterContextValue['navigate']>((to, options) => {
    const normalized = normalizePath(to)

    if (typeof window === 'undefined') {
      setPathname(normalized)
      return
    }

    if (options?.replace) {
      window.history.replaceState(null, '', normalized)
    } else {
      window.history.pushState(null, '', normalized)
    }

    setPathname(normalized)
  }, [])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    const handlePopState = () => {
      setPathname(normalizePath(window.location.pathname))
    }

    window.addEventListener('popstate', handlePopState)

    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const value = useMemo<RouterContextValue>(
    () => ({ pathname, navigate }),
    [pathname, navigate]
  )

  return <RouterContext.Provider value={value}>{children}</RouterContext.Provider>
}

type RouteProps = {
  path: string
  element: ReactNode
}

export function Route(_props: RouteProps) {
  return null
}

Route.displayName = 'Route'

type RoutesProps = {
  children: ReactNode
}

export function Routes({ children }: RoutesProps) {
  const { pathname } = useRouterContext('Routes must be used within a router context')

  let matchedElement: ReactNode = null

  Children.forEach(children, (child) => {
    if (matchedElement || !isValidElement<RouteProps>(child)) {
      return
    }

    const { path, element } = child.props

    if (matchPath(path, pathname)) {
      matchedElement = element
    }
  })

  return <>{matchedElement}</>
}

type NavLinkProps = Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href' | 'className'> & {
  to: string
  end?: boolean
  className?: string | ((state: { isActive: boolean }) => string | undefined)
  children?: ReactNode
}

export function NavLink({
  to,
  end,
  className,
  onClick,
  children,
  ...rest
}: NavLinkProps) {
  const { pathname, navigate } = useRouterContext('NavLink must be used within a router context')

  const normalizedTo = normalizePath(to)
  const normalizedPathname = normalizePath(pathname)

  const isActive = end
    ? normalizedPathname === normalizedTo
    : normalizedPathname === normalizedTo || normalizedPathname.startsWith(`${normalizedTo}/`)

  const computedClassName =
    typeof className === 'function' ? className({ isActive }) ?? '' : className ?? ''

  return (
    <a
      {...rest}
      href={normalizedTo}
      className={computedClassName}
      onClick={(event) => {
        event.preventDefault()
        navigate(normalizedTo)
        onClick?.(event)
      }}
    >
      {children}
    </a>
  )
}

export function useLocation() {
  const { pathname } = useRouterContext('useLocation must be used within a router context')

  return useMemo(
    () => ({ pathname }),
    [pathname]
  )
}

export function useNavigate() {
  const { navigate } = useRouterContext('useNavigate must be used within a router context')

  return navigate
}

function useRouterContext(message: string): RouterContextValue {
  const context = useContext(RouterContext)

  if (!context) {
    throw new Error(message)
  }

  return context
}

function matchPath(path: string, pathname: string) {
  if (path === '*') {
    return true
  }

  return normalizePath(path) === normalizePath(pathname)
}

export type { NavLinkProps, RouteProps }
