import express from 'express'
import cors from 'cors'
import dotenv from 'dotenv'
import complaintRoute from './routes/complaintRoute.js';
import adminRoute from './routes/adminRoute.js';

dotenv.config();

const app = express();
const port = process.env.PORT || 8080;

const allowedOrigins = ['http://localhost:5173'];

const corsOptions = {
    origin: function (origin, callback) {
        if (allowedOrigins.indexOf(origin) !== -1 || !origin) {
            callback(null, true);
        } else {
            callback(new Error('Not allowed by CORS'));
        }
    },
    credentials: true,
};

app.use(cors(corsOptions));
app.use(express.json());

app.use('/', complaintRoute);
app.use('/', adminRoute);

app.get('/', () => {
    res.send('Hello World!');
});

app.listen(port, () => {
    console.log(`http://localhost:${port}`);
});
