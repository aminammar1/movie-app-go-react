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
                const response = await axiosClient.get('/recommendations-ai');
                const payload = response.data?.data || [];
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