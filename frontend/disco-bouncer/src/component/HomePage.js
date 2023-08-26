import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import './HomePage.css';


function HomePage({ isLoggedIn, username }) {
  const navigate = useNavigate();

  const [students, setStudents] = useState([]);
  const [isFormVisible, setIsFormVisible] = useState(false);
  const [name, setName] = useState('');
  const [graduationYear, setGraduationYear] = useState('');
  const [selectedRoles, setSelectedRoles] = useState([]);

  const handleLogout = async () => {
    try {
      const response = await axios.post('https://discobouncer.kylrth.com/logout'); // Make a POST request to /logout
      if (response.status === 200) {
        isLoggedIn = false;
        navigate('/login');
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
    fetch('https://discobouncer.kylrth.com/api/users')
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
  
  const handleAddStudent = () => {
    setIsFormVisible(true);
  };

  const handleCloseForm = () => {
    setIsFormVisible(false);
  };
  
  const handleRoleChange = (role) => {
    if (selectedRoles.includes(role)) {
      setSelectedRoles(selectedRoles.filter(r => r !== role));
    } else {
      setSelectedRoles([...selectedRoles, role]);
    }
  };

  const handleSubmitForm = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post('https://discobouncer.kylrth.com/api/users', {
        name,
        graduationYear,
        roles: selectedRoles
      });

      if (response.status === 200) {
        // Refresh the student list or perform any other required action
        setIsFormVisible(false);
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  return (
    <div className="home-page">
      <nav className="top-navbar">
        <p className='welcome'>Welcome, {username}</p>
        <div className="nav-buttons">
          <Link to="/manage-user" className="top-navbar-link">Change Password</Link>
          <Link to="/logout" className="top-navbar-link" onClick={handleLogout}>Logout</Link>
        </div>
      </nav>
      <div className="button-row">
        <button className="student-table">Bulk Upload Students</button>
        <button className="student-table" onClick={handleAddStudent}>Add a Single Student</button>
        <button className="student-table">Bulk Decrypt Names</button>
        <button className="student-table">Single Decrypt Name</button>
      </div>
      
      {isFormVisible && (
        <div className="overlay">
          <div className="student-form">
            <h2>Add a Single Student</h2>
            <form onSubmit={handleSubmitForm}>
              <input
                type="text"
                placeholder="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
              <input
                type="text"
                placeholder="Graduation Year (YYYY)"
                value={graduationYear}
                onChange={(e) => setGraduationYear(e.target.value)}
              />
              <div className='add-student-roles'>
                <label>Roles:</label>
                <label>
                  <input
                    type="checkbox"
                    checked={selectedRoles.includes('Professor')}
                    onChange={() => handleRoleChange('Professor')}
                  />
                  Professor
                </label>
                <label>
                  <input
                    type="checkbox"
                    checked={selectedRoles.includes('TA')}
                    onChange={() => handleRoleChange('TA')}
                  />
                  TA
                </label>
                <label>
                  <input
                    type="checkbox"
                    checked={selectedRoles.includes('Student Leadership')}
                    onChange={() => handleRoleChange('Student Leadership')}
                  />
                  Student Leadership
                </label>
                <label>
                  <input
                    type="checkbox"
                    checked={selectedRoles.includes('Alumni Board')}
                    onChange={() => handleRoleChange('Alumni Board')}
                  />
                  Alumni Board
                </label>
              </div>
              <button type="submit">Submit</button>
              <button type="button" onClick={handleCloseForm}>
                Cancel
              </button>
            </form>
          </div>
        </div>
      )}
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
  );
}

export default HomePage;