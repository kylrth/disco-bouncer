import React from "react";
import { HashRouter as Router, Routes, Route } from "react-router-dom";

import './App.css';
import Login from './component/Login';
import ManageUser from './component/ManageUser';
import HomePage from './component/HomePage';

function App() {
  return (
    <div className="App">
       <Router>
        <Routes>
          <Route 
            path="/index.html" 
            element={<HomePage /> } />
          <Route 
            path="/" 
            element={<HomePage /> } />
          <Route 
            path="/home" 
            element={<HomePage /> } />
          <Route 
            path="/manage-user" 
            element={<ManageUser /> } />
          <Route 
            path="/login" 
            element={<Login /> } />
        </Routes>
      </Router>
    </div>
  );
}

export default App;
