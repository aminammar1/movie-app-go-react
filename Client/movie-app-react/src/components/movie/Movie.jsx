import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCirclePlay } from '@fortawesome/free-solid-svg-icons';
import "./Movie.css";


const Movie = ({ movie, updateMovieReview }) => {
    return (
        <div className="col-md-4 col-sm-6 mb-4" key={movie._id}>
            <Link
                to={`/stream/${movie.youtube_id}`}
                style={{ textDecoration: 'none', color: 'inherit' }}
            >
                <div className="card h-100 shadow-sm movie-card">
                    <div style={{ position: "relative" }}>
                        <img src={movie.poster_url || movie.poster_path} alt={movie.title}
                            className="card-img-top"
                            style={{
                                objectFit: "cover",
                                height: "240px",
                                width: "100%"
                            }}
                        />
                        <span className="play-icon-overlay">
                            <FontAwesomeIcon icon={faCirclePlay} />
                        </span>
                    </div>
                    <div className="card-body d-flex flex-column">
                        <h5 className="card-title">{movie.title}</h5>
                        <div className="movie-meta">
                            <span>{movie.release_year}</span>
                            {movie.genres && movie.genres.length > 0 && (
                                <span>{movie.genres[0].genre_name}</span>
                            )}
                        </div>
                    </div>
                    {movie.ranking?.ranking_name && (
                        <span className="badge bg-dark m-3 p-2" style={{ fontSize: "1rem" }}>
                            {movie.ranking.ranking_name}
                        </span>
                    )}
                    {updateMovieReview && (
                        <Button
                            variant="outline-info"
                            onClick={e => {
                                e.preventDefault();
                                updateMovieReview(movie.imdb_id);
                            }}
                            className="m-3"
                        >
                            Review
                        </Button>
                    )}
                </div>
            </Link>
        </div>
    )
}
export default Movie;