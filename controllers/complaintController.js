import axios from 'axios';
import admin from '../config/firebase.js';
import dotenv from 'dotenv';
import { v4 as uuidv4 } from 'uuid'; 

dotenv.config();

const API_URL = process.env.HUGGING_FACE_API_URL;
const API_KEY = process.env.HUGGING_FACE_API_KEY;

const bucketName = process.env.GCS_BUCKET_NAME;

async function queryImageCaption(imageData) {
    try {
        const response = await axios.post(API_URL, imageData, {
            headers: {
                'Authorization': API_KEY,
                'Content-Type': 'application/octet-stream',
            },
        });

        console.log('Hugging Face API Response:', response.data);

        if (response.data && response.data.length > 0 && response.data[0].generated_text) {
            return response.data[0].generated_text;
        } else {
            throw new Error('Invalid response format');
        }
    } catch (error) {
        console.error('Error querying image caption:', error);
        throw error;
    }
}

async function uploadImageToBucket(file) {
    if (!bucketName) {
        throw new Error('Bucket name is not configured');
    }

    const bucket = admin.storage().bucket(bucketName);
    const gcsFileName = `images/${Date.now()}-${file.originalname}`;
    const fileBuffer = file.buffer;

    const bucketFile = bucket.file(gcsFileName);

    try {
        await bucketFile.save(fileBuffer, {
            metadata: {
                contentType: file.mimetype,
            },
        });
        const [url] = await bucketFile.getSignedUrl({
            action: 'read',
            expires: '03-01-2500',
        });

        return url;
    } catch (error) {
        console.error('Error uploading to bucket:', error);
        throw error;
    }
}

export const submitPNR = async (req, res) => {
    try {
        const { pnr, subject } = req.body;
        const image = req.file;

        if (!image) {
            return res.status(400).json({ error: 'Failed to get image from request' });
        }

        console.log('Received PNR:', pnr);
        console.log('Image provided in the request');

        const imageUrl = await uploadImageToBucket(image);
        const queryGenerated = await queryImageCaption(image.buffer);

        const complaintId = uuidv4(); // Generate a unique ID for the complaint

        const complaintData = {
            id: complaintId,
            subject,
            imageUrl,
            queryGenerated,
            status: 'Pending',
            submittedAt: new Date().toISOString(),
        };

        const complaintsRef = admin.firestore().collection('complaints').doc(pnr);
        await complaintsRef.set(
            {
                complaints: admin.firestore.FieldValue.arrayUnion(complaintData),
            },
            { merge: true }
        );

        res.json({
            message: 'Complaint submitted successfully',
            pnr,
            complaintId,
            complaintData,
        });
    } catch (error) {
        console.error('Error in submit-pnr:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
};

export const getComplaints = async (req, res) => {
    try {
        const { pnr } = req.params;

        if (!pnr) {
            return res.status(400).json({ error: 'PNR is required' });
        }

        const doc = await admin.firestore().collection('complaints').doc(pnr).get();

        if (!doc.exists) {
            return res.status(404).json({ error: 'PNR not found' });
        }

        const pnrData = doc.data();

        res.json({
            pnr,
            complaints: pnrData.complaints,
        });
    } catch (error) {
        console.error('Error in get-complaints:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
};