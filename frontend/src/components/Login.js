import React, { useState, useContext } from 'react';
import { Modal, Button } from 'react-bootstrap';
import './Login.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faEnvelope, faLock } from '@fortawesome/free-solid-svg-icons';
import {broadCastFetchLoginChallenge, generateRandomString, handleConsentAndFetchToken} from "../helper/utils";
import LoggedInContext from '../context-providers/loggedin-provider/LoggedInContext';
import loginService from "../services/loginService";
import encrypt from "../helper/encrypt";

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');
  const { login } = useContext(LoggedInContext);

  const handleSuccessfulLogin = async (responseData, loginVerifier) => {
    // eslint-disable-next-line no-unused-vars
    let tokenData;
    try {
      tokenData = await handleConsentAndFetchToken(responseData.data.redirect_to, loginVerifier);
    } catch {
      // eslint-disable-next-line no-console
      console.log('something went wrong');
      return;
    }
    // eslint-disable-next-line no-console
    console.log('successfully logged in', tokenData);
    login();
  };

  const handleFailureLogin = (err) => {
    if (err?.response?.data?.errorCode === 'ERR_IDP_INVALID_CREDENTIALS') {
      console.log('Invalid email or password');
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    // Handle login logic
    console.log({ email, password });
    const code = generateRandomString();
    let loginChallengeFromResponse = '';
    try {
      loginChallengeFromResponse = await broadCastFetchLoginChallenge(code);
    } catch(err) {
      console.log(err);
      return;
    }
    loginService
        .loginUser(email, encrypt(password), loginChallengeFromResponse)
        .then((responseData) => handleSuccessfulLogin(responseData, code))
        .catch(handleFailureLogin);
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
          <h2 className="auth-title">
            Welcome Back! <span role="img" aria-label="waving hand">👋</span>
          </h2>
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

export default Login;
