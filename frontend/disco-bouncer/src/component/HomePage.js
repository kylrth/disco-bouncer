import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';
import Login from './Login';
import './HomePage.css';


function HomePage() {
  const [isLoggedIn, setIsLoggedIn] = useState(true);
  const [students, setStudents] = useState([]);

  const handleLogin = () => {
    setIsLoggedIn(true);
  };

  const handleLogout = async () => {
    try {
      const response = await axios.post('https://discobouncer.kylrth.com/logout'); // Make a POST request to /logout
      if (response.status === 200) {
        setIsLoggedIn(false);
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  useEffect(() => {
    // Fetch students data from API and set it to the state
    // Example API call (replace with your actual API endpoint):
    fetch('https://your-api-url.com/students')
      .then(response => response.json())
      .then(data => setStudents(data))
      .catch(error => console.error('Error fetching students:', error));
  }, []);

  const handleEdit = (studentId) => {
    // Handle edit action here
  };

  const handleDelete = (studentId) => {
    // Handle delete action here
  };

  const renderStudents = () => {
    return students.map(student => (
      <tr key={student.id}>
        <td>{student.name}</td>
        <td>{student.graduationYear}</td>
        <td>{student.roles.join(', ')}</td>
        <td>
          <button onClick={() => handleEdit(student.id)}>Edit</button>
          <button onClick={() => handleDelete(student.id)}>Delete</button>
        </td>
      </tr>
    ));
  };

  return (
    <div className="home-page">
    {isLoggedIn ? (
        <div>
        <nav className="top-navbar">
            <Link to="/Login" className="top-navbar-button" onClick={handleLogout}>Logout</Link>
            <Link to="/manage-user" className="top-navbar-button">Change Password</Link>
        </nav>
        <div className="button-row">
            <button className="student-table">Bulk Upload Students</button>
            <button className="student-table">Add a Single Student</button>
            <button className="student-table">Bulk Decrypt Names</button>
            <button className="student-table">Single Decrypt Name</button>
        </div>
        <div className="student-table-container">
            <table className="student-table">
            <thead>
                <tr>
                <th>Name/ID</th>
                <th>Graduation Year</th>
                <th>Roles</th>
                <th>Edit/Delete</th>
                </tr>
            </thead>
            <tbody>
                {renderStudents()}
            </tbody>
            </table>
        </div>
        </div>
    ) : (
        <Login onLogin={handleLogin} />
    )}
    </div>
  );
}

export default HomePage;