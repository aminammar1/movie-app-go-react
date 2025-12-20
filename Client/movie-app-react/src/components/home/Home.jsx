import { useState, useEffect } from 'react';
import axiosClient from '../../api/axiosConfig'
import Movies from '../movies/Movies';
import Spinner from '../spinner/Spinner';
import './home.css';

const Home = ({ updateMovieReview }) => {
    const [movies, setMovies] = useState([]);
    const [featured, setFeatured] = useState(null);
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
                setFeatured(payload[0] || null);
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
            <div className="page-shell">
                {featured && (
                    <section className="hero" style={{ backgroundImage: `linear-gradient(120deg, rgba(20, 23, 38, 0.9), rgba(11, 13, 23, 0.75)), url(${featured.poster_url})` }}>
                        <div className="hero-content">
                            <div className="pill-row">
                                {featured.genres?.slice(0, 2).map(g => (
                                    <span key={g.genre_id} className="pill">{g.genre_name}</span>
                                ))}
                            </div>
                            <h1 className="hero-title">{featured.title}</h1>
                            <p className="hero-subtitle">{featured.description?.slice(0, 180)}...</p>
                            <div className="cta-row">
                                <button className="btn-primary-neo" onClick={() => updateMovieReview(featured.imdb_id)}>Admin review</button>
                                <button className="btn-ghost" onClick={() => window.location.href = `/stream/${featured.youtube_id}`}>Watch trailer</button>
                            </div>
                        </div>
                    </section>
                )}
                <h2 className="section-title">Popular Now</h2>
                {error && <div className="alert alert-danger">{error}</div>}
                {loading ? (
                    <Spinner />
                ) : (
                    <Movies movies={movies} updateMovieReview={updateMovieReview} message={message} />
                )}
            </div>
        </>

    );

};

export default Home;
