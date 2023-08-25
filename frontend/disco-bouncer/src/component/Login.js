import React, { useState } from 'react';
import axios from 'axios';
import './Login.css';

function Login({ onLogin }) {
  const [input_username, setUsername] = useState('');
  const [input_password, setPassword] = useState('');

  const handleLogin = async () => {
    try {
      const response = await axios.post('https://discobouncer.kylrth.com/login', {
        username: input_username,
        password: input_password
      });
  
      if (response.status === 200) {
        // Call the callback function passed as prop
        onLogin();
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  return (
    <div className="login-page">
        <div className="logo-container">
            <img src="./logo.png" alt="Disco Bouncer Logo" className="logo" />
        </div>
        <br />
        <br />
        <div className='login-box'>
            <input
                className='input-field'
                type="text"
                value={input_username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Username"
            />
            <input
                className='input-field'
                type="password"
                value={input_password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
            />
            <button onClick={handleLogin}>Login</button>
        </div>
    </div>
  );
}

export default Login;