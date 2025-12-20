import { useState, useEffect } from 'react';
import axiosClient from '../../api/axiosConfig'
import Movies from '../movies/Movies';
import Spinner from '../spinner/Spinner';

const Home = ({ updateMovieReview }) => {
    const [movies, setMovies] = useState([]);
    const [loading, setLoading] = useState(false)
    const [message, setMessage] = useState();
    const [error, setError] = useState();

    useEffect(() => {
        const fetchMovies = async () => {
            setLoading(true);
            setMessage("");
            try {
                const response = await axiosClient.get('/movies');
                const payload = response.data?.data || [];
                setMovies(payload);
                if (payload.length === 0) {
                    setMessage('There are currently no movies available')
                }

            } catch (error) {
                console.error('Error fetching movies:', error)
                setError('Unable to load movies. Please try again later.')
            } finally {
                setLoading(false)
            }
        }
        fetchMovies();
    }, []);

    return (
        <>
            <h2>Movie App</h2>
            {error && <div className="alert alert-danger">{error}</div>}
            {loading ? (
                <Spinner />
            ) : (
                <Movies movies={movies} updateMovieReview={updateMovieReview} message={message} />
            )}
        </>

    );

};

export default Home;
