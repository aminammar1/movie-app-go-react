import { useEffect } from 'react';
import axios from 'axios';

import useAuth from './useAuth';

const apiUrl = import.meta.env.VITE_API_URL || window._ENV_?.VITE_API_URL || window._ENV_?.API_URL;

const useAxios = () => {

    const axiosAuth = axios.create({
        baseURL: apiUrl,
        withCredentials: true
    });


    const { auth, setAuth } = useAuth();

    let isRefreshing = false;
    let failedQueue = [];

    // Function to process the failed requests queue
    const processQueue = (error, response = null) => {
        failedQueue.forEach(prom => {
            if (error) {
                prom.reject(error);
            } else {
                prom.resolve(response);
            }
        });

        failedQueue = [];
    };

    useEffect(() => {

        axiosAuth.interceptors.response.use(
            response => response,
            async error => {
                console.log('Interceptor caught error:', error);
                const originalRequest = error.config;

                if (originalRequest.url.includes('/refresh-token') && error.response.status === 401) {
                    // the refresh token itself is invalid/expired
                    console.error('Refresh token is invalid or expired.');
                    return Promise.reject(error); // fail directly, no retry
                }

                if (error.response && error.response.status === 401 && !originalRequest._retry) {

                    if (isRefreshing) {
                        return new Promise((resolve, reject) => {
                            failedQueue.push({ resolve, reject });
                        })
                            .then(() => axiosAuth(originalRequest))
                            .catch(err => Promise.reject(err));
                    }

                    originalRequest._retry = true;
                    isRefreshing = true;

                    return new Promise((resolve, reject) => {
                        axiosAuth
                            .post('/refresh-token')
                            .then(() => {

                                processQueue(null);

                                axiosAuth(originalRequest)
                                    .then(resolve)
                                    .catch(reject);

                            })
                            .catch(refreshError => {

                                processQueue(refreshError, null);

                                localStorage.removeItem('user');
                                setAuth(null); // Clear auth state
                                reject(refreshError); // fail the original promise chain
                            })
                            .finally(() => {
                                isRefreshing = false;
                            });
                    });
                }

                return Promise.reject(error);
            }
        );

    }, [auth]);

    return axiosAuth;
}

export default useAxios;