import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from './AuthProvider';
import './HomePage.css';


function HomePage() {
  const { authenticated, cookies, logout, setAuthenticated, setCookies } = useAuth();
  const [isLoading, setLoading] = useState(true); // Initial loading state
  const [isRestoringAuth, setRestoringAuth] = useState(true); // Flag for restoring auth


  const [userData, setUserData] = useState('');
  const [students, setStudents] = useState([]);
  const [studentId, setStudentId] = useState([]);
  const [encryptionKey, setEncryptionKey] = useState([]);
  const [isSingleUploadFormVisible, setIsSingleUploadFormVisible] = useState(false);
  const [isBulkUploadFormVisible, setIsBulkUploadFormVisible] = useState(false);
  const [isBulkDecryptFormVisible, setIsBulkDecryptFormVisible] = useState(false);
  const [isSingleDecryptFormVisible, setIsSingleDecryptFormVisible] = useState(false);
  const [name, setName] = useState('');
  const [graduationYear, setGraduationYear] = useState('');
  const [selectedRoles, setSelectedRoles] = useState([]);

  const navigate = useNavigate();

  const fetchStudents = async () => {
    try {
      const response = await axios.get('https://discobouncer.kylrth.com/api/users', {
        withCredentials: true, // Include cookies
      });
      
      const data = response.data;
      setStudents(data);
      
      console.log("getting students");
      console.log(data); // Use 'data' instead of 'students' here
    } catch (error) {
      console.error('Error fetching students:', error);
    }
  };

  const fetchUserData = async () => {
    try {
      const response = await axios.get('https://discobouncer.kylrth.com/api/users', {
        withCredentials: true, // Include cookies
        headers: {
          Cookie: cookies, // Set the 'Cookie' header using the 'cookies' variable
        },
      });
      
      const data = response.data;
      setUserData(data);
    } catch (error) {
      console.error('Error fetching user data:', error);
    }
  };

  useEffect(() => {
    const savedAuthenticated = localStorage.getItem('authenticated');
    const savedCookies = localStorage.getItem('cookies');

    if (savedAuthenticated && savedCookies) {
      setAuthenticated(JSON.parse(savedAuthenticated));
      setCookies(savedCookies);
    }

    setLoading(false); // Set loading to false once authentication is restored
    setRestoringAuth(false); // Set restoring auth to false

    if (!authenticated && !isRestoringAuth) {
      navigate('/login');
    } else {
      if (authenticated && cookies) {
        fetchUserData();
        fetchStudents();
      }
    }
  }, [authenticated, cookies, navigate, fetchUserData, fetchStudents]);

  const handleLogout = async () => {
    try {
      const response = await axios.post('https://discobouncer.kylrth.com/logout'); // Make a POST request to /logout
      if (response.status === 200) {
        logout();
        navigate('/login');
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

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
  
  const handleRoleChange = (role) => {
    if (selectedRoles.includes(role)) {
      setSelectedRoles(selectedRoles.filter(r => r !== role));
    } else {
      setSelectedRoles([...selectedRoles, role]);
    }
  };

  const handleSubmitSingleUploadForm = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post('https://discobouncer.kylrth.com/api/users', {
        withCredentials: true, // Include cookies
        headers: {
          Cookie: cookies, // Set the 'Cookie' header using the 'cookies' variable
        },
        name,
        graduationYear,
        roles: selectedRoles
      });

      if (response.status === 200) {
        // Refresh the student list or perform any other required action
        setIsSingleUploadFormVisible(false);
        fetchStudents();
        renderStudents();
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  const handleBulkUpload = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append('csv', e.target.files[0]);

    try {
      const response = await axios.post('https://discobouncer.kylrth.com/api/users', formData, { //accepts a json of a user
        withCredentials: true, // Include cookies
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });

      if (response.status === 200) {
        // Refresh the student list or perform any other required action
        setIsBulkUploadFormVisible(false)();
        fetchStudents();
        renderStudents();
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  const handleBulkDecrypt = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append('csv', e.target.files[0]);

    try {
      const response = await axios.post('https://discobouncer.kylrth.com/api/bulk-decrypt', formData, {
        withCredentials: true, // Include cookies
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });

      if (response.status === 200) {
        // Process the decrypted data (e.g., display in a modal or update state)
        setIsBulkDecryptFormVisible(false)();
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  const handleSingleDecrypt = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append('studentId', e.target[0].value);
    formData.append('encryptionCode', e.target[1].value);

    try {
      const response = await axios.post('https://discobouncer.kylrth.com/api/single-decrypt', formData,{
        withCredentials: true, // Include cookies
      });

      if (response.status === 200) {
        // Process the decrypted data (e.g., display in a modal or update state)
        setIsSingleDecryptFormVisible(false)();
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  if (isLoading) {
    return <div>Loading...</div>; // Show a loading indicator during the initial loading phase
  }

  return (
    <div className="home-page">
      <nav className="top-navbar">
        <p className='welcome'>Welcome, {userData.username}</p>
        <div className="nav-buttons">
          <Link to="/manage-user" className="top-navbar-link">Change Password</Link>
          <Link to="/logout" className="top-navbar-link" onClick={handleLogout}>Logout</Link>
        </div>
      </nav>
      <div className="button-row">
        <button className="student-table" onClick={() => setIsBulkUploadFormVisible(true)}>Bulk Upload Students</button>
        <button className="student-table" onClick={() => setIsSingleUploadFormVisible(true)}>Add a Single Student</button>
        <button className="student-table" onClick={() => setIsBulkDecryptFormVisible(true)}>Bulk Decrypt Names</button>
        <button className="student-table" onClick={() => setIsSingleDecryptFormVisible(true)}>Single Decrypt Name</button>
      </div>
      
      {isSingleUploadFormVisible && (
        <div className="overlay">
          <div className="student-form">
            <h2>Add a Single Student</h2>
            <form onSubmit={handleSubmitSingleUploadForm}>
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
              <button type="button" onClick={() => setIsSingleUploadFormVisible(false)}>
                Cancel
              </button>
            </form>
          </div>
        </div>
      )}

      {isBulkUploadFormVisible && (
        <div className="overlay">
          <div className="student-form">
            <h2>Add Many Students via CSV Upload</h2>
            <a href="./sample-bulk-upload.csv" download>Download Sample CSV Template</a>
            <br />
            <br />
            <input
              type="file"
              accept=".csv"
              id="bulk-upload"
              style={{ display: 'none' }}
              onChange={handleBulkUpload}
            />
            <label htmlFor="bulk-upload" className="student-table">
              Bulk Upload Students
            </label>
            <button onClick={() => setIsBulkUploadFormVisible(false)}>Cancel</button>
          </div>
        </div>
            
        
      )}

      {isBulkDecryptFormVisible && (
        <div className="overlay">
          <div className="student-form">
            <h2>Decrypt Many Students via CSV Upload</h2>
            <p>Upload a CSV with student IDs and encryption keys to decrypt IDs into student names.</p>
            <a href="./sample-bulk-decrypt.csv" download>Download Sample CSV Template</a>
            <br />
            <br />
            <input
              type="file"
              accept=".csv"
              id="bulk-decrypt"
              style={{ display: 'none' }}
              onChange={handleBulkDecrypt}
            />
            <label htmlFor="bulk-decrypt" className="student-table">
              Upload CSV to Bulk Decrypt Students
            </label>
            <button onClick={() => setIsBulkDecryptFormVisible(false)}>Cancel</button>
          </div>
        </div>
            
        
      )}

      {isSingleDecryptFormVisible && (
        <div className="overlay">
          <div className="student-form">
            <h2>Decrypt a Single Student Name</h2>
            
            <form onSubmit={handleSingleDecrypt}>
              <input
                type="text"
                placeholder="Student ID"
                value={studentId}
                onChange={(e) => setStudentId(e.target.value)}
              />
              <input
                type="text"
                placeholder="Encryption Key"
                value={encryptionKey}
                onChange={(e) => setEncryptionKey(e.target.value)}
              />
              <button type="submit">Submit</button>
              <button onClick={() => setIsSingleDecryptFormVisible(false)}>Cancel</button>
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