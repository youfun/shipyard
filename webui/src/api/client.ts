/**
 * Layer 1: Base HTTP Client
 * Axios instance with baseURL configuration and authentication
 */
import axios from 'axios';

const TOKEN_KEY = 'shipyard_access_token';

function getAuthToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(TOKEN_KEY);
  }
  return null;
}

const apiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth interceptor
apiClient.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle response errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid, clear it
      localStorage.removeItem(TOKEN_KEY);
      // Don't redirect here, let the layout handle it with proper redirect params
      // The AuthContext will detect the missing token and trigger the auth flow
    }
    return Promise.reject(error);
  }
);


// 全局响应拦截器 (removed duplicate 401 handling)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;

    if (status >= 500) {
      // 5xx 服务器错误
      console.error('[API] Server error:', error.response?.data || error.message);
      // 建议替换为 Toast 组件提示
      alert('服务器开小差了 (500)，请稍后再试。');
    } else if (status !== 401) {
      // 其他错误 (400, 404 等) - 401 already handled above
      console.error('[API] Request failed:', error.response?.data || error.message);
    }

    // 将错误抛回给调用者，以便组件在必要时进行特定处理（如表单报错）
    return Promise.reject(error);
  }
);




export default apiClient;
