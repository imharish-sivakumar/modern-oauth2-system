import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Login from './components/Login';
import Signup from './components/Signup';
import VerifyAccount from './components/VerifyAccount';
import LoggedInProvider from "./context-providers/loggedin-provider/LoggedInProvider";

const App = () => {
  return (
    <Router>
      <LoggedInProvider>
        <div>
          <Routes>
            <Route path="/" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
            <Route path="/verify" element={<VerifyAccount />} /> {/* New route for verification */}
          </Routes>
        </div>
      </LoggedInProvider>
    </Router>
  );
};

export default App;
