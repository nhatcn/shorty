import React, { useState } from 'react';
import { setCookie } from '../utils/cookieUltil';

const API_BASE_URL = process.env.BE_URL ;

const authAPI = {
  register: async (username, password) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Wrong username or password');
        }
        throw new Error('Network error or server issue');
      }
      
      return response;
    } catch (error) {
      if (error.message === 'Wrong username or password') {
        throw error;
      }
      throw new Error('Network error. Please check your connection.');
    }
  },

  login: async (username, password) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Wrong username or password');
        }
        throw new Error('Network error or server issue');
      }
      
      const data = await response.json();
      setCookie('userId', data.userId);
      localStorage.setItem('token', data.token);
      return data;
    } catch (error) {
      if (error.message === 'Wrong username or password') {
        throw error;
      }
      throw new Error('Network error. Please check your connection.');
    }
  }
};

export default function Auth({ onLoginSuccess = () => {} }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loginUsername, setLoginUsername] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [isSignUp, setIsSignUp] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);

  const handleAuth = async () => {
    if (!loginUsername || !loginPassword) {
      setError('Please fill in all fields');
      return;
    }

    setLoading(true);
    setError('');
    setSuccess('');

    try {
      if (isSignUp) {
        await authAPI.register(loginUsername, loginPassword);
        setIsSignUp(false);
        setSuccess('Account created successfully! Please sign in.');
        setLoginPassword('');
      } else {
        const data = await authAPI.login(loginUsername, loginPassword);
        const userData = { 
          username: loginUsername, 
          id: Date.now(),
          token: data.token 
        };
        
        setIsLoggedIn(true);
        setCurrentUser(userData);
        onLoginSuccess(userData);
        
        // Redirect to /home
        window.location.href = '/home';
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleMode = () => {
    setIsSignUp(!isSignUp);
    setError('');
    setSuccess('');
    setLoginPassword('');
  };

  return (
    <div className="min-vh-100 d-flex align-items-center justify-content-center" style={{ backgroundColor: '#f8f9fa' }}>
      {isLoggedIn ? (
        <div className="card shadow-lg border-0" style={{ maxWidth: '450px', width: '100%' }}>
          <div className="card-body p-5 text-center">
            <div className="alert alert-success mb-4" role="alert">
              <h4 className="alert-heading">Login Successful!</h4>
              <p className="mb-0">Welcome back, {currentUser?.username}!</p>
            </div>
            <p className="text-muted">You are now logged in.</p>
          </div>
        </div>
      ) : (
        <div className="card shadow-lg border-0" style={{ maxWidth: '450px', width: '100%' }}>
          <div className="card-body p-5">
            <div className="text-center mb-4">
              <div className="d-inline-flex align-items-center justify-content-center bg-dark text-white rounded mb-3" 
                   style={{ width: '48px', height: '48px', fontWeight: '700', fontSize: '24px' }}>
                S
              </div>
              <h2 className="fw-bold mb-2">{isSignUp ? 'Create Account' : 'Welcome Back'}</h2>
              <p className="text-muted">
                {isSignUp ? 'Sign up to start shortening URLs' : 'Sign in to your account'}
              </p>
            </div>

            {error && (
              <div className="alert alert-danger" role="alert">
                {error}
              </div>
            )}

            {success && (
              <div className="alert alert-success" role="alert">
                {success}
              </div>
            )}

            <div className="mb-3">
              <label className="form-label">Username</label>
              <input
                type="text"
                className="form-control form-control-lg"
                placeholder="Enter username"
                value={loginUsername}
                onChange={(e) => setLoginUsername(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleAuth()}
              />
            </div>

            <div className="mb-4">
              <label className="form-label">Password</label>
              <input
                type="password"
                className="form-control form-control-lg"
                placeholder="Enter your password"
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleAuth()}
              />
            </div>

            <button 
              onClick={handleAuth}
              className="btn btn-dark btn-lg w-100 mb-3"
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                  {isSignUp ? 'Creating account...' : 'Signing in...'}
                </>
              ) : (
                isSignUp ? 'Sign Up' : 'Sign In'
              )}
            </button>

            <div className="text-center">
              <button 
                className="btn btn-link text-decoration-none"
                onClick={handleToggleMode}
              >
                {isSignUp ? 'Already have an account? Sign in' : "Don't have an account? Sign up"}
              </button>
            </div>
          </div>
        </div>
      )}

      <link 
        href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" 
        rel="stylesheet" 
        crossOrigin="anonymous"
      />
    </div>
  );
}