import React, {useEffect} from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Login from './components/Login';
import Signup from './components/Signup';
import VerifyAccount from './components/VerifyAccount';
import LoggedInProvider from "./context-providers/loggedin-provider/LoggedInProvider";
import LoggedInContext from "./context-providers/loggedin-provider/LoggedInContext";
import Home from "./components/Home";

const App = () => {
  return (
    <Router>
      <LoggedInProvider>
        <div>
          <Routes>
            <Route index element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
            <Route path="/verify" element={<VerifyAccount />} /> {/* New route for verification */}
          </Routes>
        </div>
      </LoggedInProvider>
    </Router>
  );
};

export default App;
