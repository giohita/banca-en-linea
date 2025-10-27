import axios from 'axios';

// Configuración base de Axios
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Interceptor para agregar el token de autorización
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Interceptor para manejar respuestas y errores
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expirado o inválido
      localStorage.removeItem('authToken');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Servicios de autenticación
export const authService = {
  register: async (userData) => {
    const response = await api.post('/auth/register', userData);
    return response.data;
  },

  login: async (credentials) => {
    console.log('🌐 API: Enviando petición de login con:', credentials);
    console.log('🌐 API: URL del backend:', API_BASE_URL);
    
    try {
      const response = await api.post('/auth/login', credentials);
      console.log('🌐 API: Respuesta completa del servidor:', response);
      console.log('🌐 API: Status:', response.status);
      console.log('🌐 API: Data:', response.data);
      
      if (response.data.token) {
        localStorage.setItem('authToken', response.data.token);
        localStorage.setItem('user', JSON.stringify(response.data.user));
        console.log('🌐 API: Token y usuario guardados en localStorage');
      }
      return response.data;
    } catch (error) {
      console.error('🌐 API: Error en petición de login:', error);
      console.error('🌐 API: Error response:', error.response);
      console.error('🌐 API: Error status:', error.response?.status);
      console.error('🌐 API: Error data:', error.response?.data);
      throw error;
    }
  },

  logout: async () => {
    try {
      await api.post('/auth/logout');
    } finally {
      localStorage.removeItem('authToken');
      localStorage.removeItem('user');
    }
  },

  getCurrentUser: async () => {
    const response = await api.get('/auth/me');
    return response.data;
  },

  isAuthenticated: () => {
    return !!localStorage.getItem('authToken');
  },

  getStoredUser: () => {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  }
};

// Servicios de usuarios
export const userService = {
  getUsers: async (limit = 10, offset = 0) => {
    const response = await api.get(`/users?limit=${limit}&offset=${offset}`);
    return response.data;
  },

  getUser: async (userId) => {
    const response = await api.get(`/users/${userId}`);
    return response.data;
  },

  getUserBalance: async (userId) => {
    const response = await api.get(`/users/${userId}/balance`);
    return response.data;
  },

  createUser: async (userData) => {
    const response = await api.post('/users', userData);
    return response.data;
  }
};

// Servicios de transacciones
export const transactionService = {
  deposit: async (userId, amount) => {
    const response = await api.post(`/users/${userId}/deposit`, { amount });
    return response.data;
  },

  withdraw: async (userId, amount) => {
    const response = await api.post(`/users/${userId}/withdraw`, { amount });
    return response.data;
  },

  transfer: async (fromUserId, toUserId, amount) => {
    const response = await api.post('/transfer', {
      from_user_id: fromUserId,
      to_user_id: toUserId,
      amount
    });
    return response.data;
  }
};

// Servicio de salud
export const healthService = {
  check: async () => {
    const response = await api.get('/health');
    return response.data;
  }
};

export default api;