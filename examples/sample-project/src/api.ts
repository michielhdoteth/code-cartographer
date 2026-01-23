import { AuthManager } from './auth'

interface Request {
  method: string
  path: string
  headers: Record<string, string>
  body?: any
}

interface Response {
  status: number
  data: any
  headers?: Record<string, string>
}

class APIHandler {
  private auth: AuthManager
  private routes: Map<string, Function> = new Map()

  constructor(auth: AuthManager) {
    this.auth = auth
    this.initializeRoutes()
  }

  private initializeRoutes(): void {
    this.routes.set('GET /users', this.getUsers.bind(this))
    this.routes.set('POST /users', this.createUser.bind(this))
    this.routes.set('GET /profile', this.getProfile.bind(this))
  }

  async handle(request: Request): Promise<Response> {
    const routeKey = `${request.method} ${request.path}`
    const handler = this.routes.get(routeKey)

    if (!this.auth.isAuthenticated()) {
      return { status: 401, data: { error: 'Unauthorized' } }
    }

    if (handler) {
      return await handler(request)
    }

    return { status: 404, data: { error: 'Not found' } }
  }

  private async getUsers(request: Request): Promise<Response> {
    return { status: 200, data: { users: [] } }
  }

  private async createUser(request: Request): Promise<Response> {
    return { status: 201, data: { id: 1, ...request.body } }
  }

  private async getProfile(request: Request): Promise<Response> {
    return { status: 200, data: { profile: this.auth.getToken() } }
  }
}

export { APIHandler, Request, Response }
