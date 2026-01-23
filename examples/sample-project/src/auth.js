// Authentication Module
class AuthManager {
  constructor(config) {
    this.config = config
    this.token = null
    this.user = null
  }

  async login(email, password) {
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })

      if (response.ok) {
        const data = await response.json()
        this.token = data.token
        this.user = data.user
        return { success: true, user: this.user }
      }
    } catch (error) {
      console.error('Login failed:', error)
      return { success: false, error: error.message }
    }
  }

  logout() {
    this.token = null
    this.user = null
  }

  isAuthenticated() {
    return this.token !== null
  }

  getToken() {
    return this.token
  }
}

module.exports = AuthManager
