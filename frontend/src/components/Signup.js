import React, { useState } from 'react';
import { Modal, Button } from 'react-bootstrap';
import './Signup.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faEnvelope, faLock } from '@fortawesome/free-solid-svg-icons';
import loginService from "../services/loginService";
import encrypt from "../helper/encrypt";

const Signup = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');
  const [showVerificationModal, setShowVerificationModal] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    try {
      const encryptedPassword = encrypt(password);
      const response = await loginService.registerUser(email, encryptedPassword);

      if (response.status === 200) {
        setShowVerificationModal(true);
      } else {
        console.error('Registration failed:', response.data);
        alert('Registration failed. Please try again.');
      }
    } catch (error) {
      console.error('Error during registration:', error);
      alert('An error occurred during registration. Please try again.');
    }
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
          <h2 className="auth-title">Become a Member Today!</h2>
          <div className="form-group">
            <label>Email</label>
            <div className="input-with-icon">
              <FontAwesomeIcon icon={faEnvelope} />
              <input
                type="email"
                className="form-control"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>
          <div className="form-group">
            <label>Password</label>
            <div className="input-with-icon">
              <FontAwesomeIcon icon={faLock} />
              <input
                type="password"
                className="form-control"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
          </div>
          <button type="submit" className="btn btn-primary btn-block">
            Sign up
          </button>
          <div className="login-reset-container">
            <p className="register-link">
              Already a member? <a href="/">Login here</a>
            </p>
            <p className="reset-password-link">
              Forgot your password?{' '}
              <a
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  setShowForgotPassword(true);
                }}
              >
                Reset here
              </a>
            </p>
          </div>
        </form>
      </div>

      {/* Verification Modal */}
      <Modal
        show={showVerificationModal}
        onHide={() => setShowVerificationModal(false)}
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>Check Your Email</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>Account created successfully! Please check your email to verify your account.</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={() => setShowVerificationModal(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>

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
              <div className="input-with-icon">
                <FontAwesomeIcon icon={faEnvelope} />
                <input
                  type="email"
                  className="form-control"
                  placeholder="Enter your email"
                  value={forgotEmail}
                  onChange={(e) => setForgotEmail(e.target.value)}
                  required
                />
              </div>
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

export default Signup;
