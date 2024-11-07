import React, { useState } from 'react';
import { Modal, Button } from 'react-bootstrap';
import './Login.css';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');

  const handleSubmit = (event) => {
    event.preventDefault();
    // Handle login logic
    console.log({ email, password });
  };

  const handleForgotPasswordSubmit = () => {
    // Handle forgot password logic
    console.log('Reset password for:', forgotEmail);
    setShowForgotPassword(false);
  };

  return (
    <div className="auth-container">
      <div className="auth-form-container">
        <form onSubmit={handleSubmit} className="auth-form">
          <h2 className="auth-title">Welcome back 👋</h2>
          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              className="form-control"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input
              type="password"
              className="form-control"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <div className="options-container">
            <div>
              <input type="checkbox" id="keepLoggedIn" />
              <label htmlFor="keepLoggedIn"> Keep me logged in</label>
            </div>
            <a
              href="#"
              className="forgot-password-link"
              onClick={(e) => {
                e.preventDefault();
                setShowForgotPassword(true);
              }}
            >
              Forgot password?
            </a>
          </div>
          <button type="submit" className="btn btn-primary btn-block">
            Sign in
          </button>
          <p className="register-link">
            Not a member yet? <a href="/signup">Register now</a>
          </p>
        </form>
      </div>

      {/* Forgot Password Modal */}
      <Modal
        show={showForgotPassword}
        onHide={() => setShowForgotPassword(false)}
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>Forgot Password</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <form>
            <div className="form-group">
              <label>Email</label>
              <input
                type="email"
                className="form-control"
                placeholder="Enter your email"
                value={forgotEmail}
                onChange={(e) => setForgotEmail(e.target.value)}
                required
              />
            </div>
          </form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={handleForgotPasswordSubmit}>
            Reset Password
          </Button>
          <Button variant="secondary" onClick={() => setShowForgotPassword(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default Login;
