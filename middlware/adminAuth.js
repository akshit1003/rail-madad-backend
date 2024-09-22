import { db } from '../config/firebase.js';

export const adminAuth = async (req, res, next) => {
    const { idToken } = req.body; 

    try {
        const decodedToken = await admin.auth().verifyIdToken(idToken);
        const userEmail = decodedToken.email;

        const adminDoc = await db.collection('admin').where('email', '==', userEmail).get();

        if (adminDoc.empty) {
            return res.status(403).json({ message: 'Access Denied: Not an admin' });
        }
        
        const email = decodedToken.email;
        req.email = email;
        
        next();
    } catch (error) {
        console.error('Error in checking admin email:', error);
        res.status(500).json({ message: 'Internal server error', error: error.message });
    }
};
