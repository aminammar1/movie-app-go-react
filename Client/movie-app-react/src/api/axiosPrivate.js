import axios from "axios";

const api_url = import.meta.env.VITE_API_URL;

const axiosPrivate = axios.create({
  baseURL: api_url,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

export default axiosPrivate;