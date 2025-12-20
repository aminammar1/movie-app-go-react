import useAxios from '../../hooks/useAxios';
import { useEffect, useState } from 'react';
import Movies from '../movies/Movies';
import Spinner from '../spinner/Spinner';

const Recommandation = () => {
    const [movies, setMovies] = useState([]);
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState();
    const axiosClient = useAxios();

    useEffect(() => {
        const fetchRecommendedMovies = async () => {
            setLoading(true);
            setMessage("");

            try {
                const base = await axiosClient.get('/recommendatedmovies');
                let payload = base.data?.data || [];

                if (!payload.length) {
                    const controller = new AbortController();
                    const timer = setTimeout(() => controller.abort(), 7000);
                    try {
                        const aiResp = await axiosClient.get('/recommendations-ai', { signal: controller.signal });
                        payload = aiResp.data?.data || [];
                    } catch (err) {
                        console.error('AI recommendations fallback timed out or failed');
                    } finally {
                        clearTimeout(timer);
                    }
                }

                setMovies(payload);
            } catch (error) {
                console.error("Error fetching recommended movies:", error)
            } finally {
                setLoading(false);
            }

        }
        fetchRecommendedMovies();
    }, [])

    return (
        <>
            {loading ? (
                <Spinner />
            ) : (
                <Movies movies={movies} message={message} />
            )}
        </>
    )

}
export default Recommandation