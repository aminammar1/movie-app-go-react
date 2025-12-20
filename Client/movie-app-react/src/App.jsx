import './App.css'
import Home from './components/home/Home';
import Recommandation from './components/recommended/Recommandation';
import Review from './components/review/Review';
import Header from './components/header/Header';
import Register from './components/register/Register';
import Login from './components/login/Login';
import Layout from './components/Layout.jsx';
import RequiredAuth from './components/Auth.jsx';
import axiosClient from './api/axiosConfig';
import useAuth from './hooks/useAuth';
import StreamMovie from './components/stream/StreamingMovie';

import { Route, Routes, useNavigate } from 'react-router-dom'

function App() {

  const navigate = useNavigate();
  const { auth, setAuth } = useAuth();


  const updateMovieReview = (imdb_id) => {
    navigate(`/review/${imdb_id}`);
  };

  const handleLogout = async () => {

    try {
      const response = await axiosClient.post("/logout", { user_id: auth.user_id });
      console.log(response.data);
      setAuth(null);
      console.log('User logged out');

    } catch (error) {
      console.error('Error logging out:', error);
    }

  };

  return (
    <>
      <Header handleLogout={handleLogout} />
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<Home updateMovieReview={updateMovieReview} />} />
          <Route path="register" element={<Register />} />
          <Route path="login" element={<Login />} />
          <Route element={<RequiredAuth />}>
            <Route path="recommended" element={<Recommandation />} />
            <Route path="review/:imdb_id" element={<Review />} />
            <Route path="stream/:yt_id" element={<StreamMovie />} />
          </Route>
        </Route>
      </Routes>

    </>
  )
}

export default App