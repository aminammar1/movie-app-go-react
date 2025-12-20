import axios from "axios";

const apiUrl = import.meta.env.VITE_API_URL || window._ENV_?.VITE_API_URL || window._ENV_?.API_URL;

const axiosPrivate = axios.create({
  baseURL: apiUrl,
  headers: {
    "Content-Type": "application/json"
  },
  withCredentials: true
});

export default axiosPrivate;