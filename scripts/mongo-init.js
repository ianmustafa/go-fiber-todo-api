// MongoDB initialization script
// This script runs when the MongoDB container starts for the first time

// Switch to the go_fiber database
db = db.getSiblingDB('go_fiber');

// Create collections with validation
db.createCollection('users', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['username', 'password_hash', 'created_at', 'updated_at'],
      properties: {
        _id: {
          bsonType: 'string',
          description: 'ULID string identifier'
        },
        username: {
          bsonType: 'string',
          minLength: 3,
          maxLength: 50,
          description: 'Username must be a string between 3-50 characters'
        },
        password_hash: {
          bsonType: 'string',
          description: 'Hashed password'
        },
        email: {
          bsonType: ['string', 'null'],
          pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
          description: 'Valid email address or null'
        },
        image: {
          bsonType: ['string', 'null'],
          description: 'URL to user avatar image or null'
        },
        created_at: {
          bsonType: 'date',
          description: 'Creation timestamp'
        },
        updated_at: {
          bsonType: 'date',
          description: 'Last update timestamp'
        }
      }
    }
  }
});

db.createCollection('todos', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'title', 'status', 'created_at', 'updated_at'],
      properties: {
        _id: {
          bsonType: 'string',
          description: 'ULID string identifier'
        },
        user_id: {
          bsonType: 'string',
          description: 'Reference to user ULID'
        },
        title: {
          bsonType: 'string',
          minLength: 1,
          maxLength: 200,
          description: 'Todo title must be 1-200 characters'
        },
        description: {
          bsonType: ['string', 'null'],
          description: 'Optional todo description'
        },
        status: {
          bsonType: 'string',
          enum: ['pending', 'in_progress', 'completed'],
          description: 'Todo status'
        },
        priority: {
          bsonType: ['string', 'null'],
          enum: ['low', 'medium', 'high', null],
          description: 'Todo priority'
        },
        due_date: {
          bsonType: ['date', 'null'],
          description: 'Optional due date'
        },
        created_at: {
          bsonType: 'date',
          description: 'Creation timestamp'
        },
        updated_at: {
          bsonType: 'date',
          description: 'Last update timestamp'
        }
      }
    }
  }
});

// Create indexes for users collection
db.users.createIndex({ 'username': 1 }, { unique: true });
db.users.createIndex({ 'email': 1 }, { unique: true, sparse: true });
db.users.createIndex({ 'created_at': 1 });

// Create indexes for todos collection
db.todos.createIndex({ 'user_id': 1 });
db.todos.createIndex({ 'user_id': 1, 'status': 1 });
db.todos.createIndex({ 'user_id': 1, 'priority': 1 });
db.todos.createIndex({ 'user_id': 1, 'due_date': 1 });
db.todos.createIndex({ 'status': 1 });
db.todos.createIndex({ 'due_date': 1 });
db.todos.createIndex({ 'created_at': 1 });

// Create text index for search functionality
db.todos.createIndex({ 
  'title': 'text', 
  'description': 'text' 
}, {
  weights: {
    'title': 10,
    'description': 5
  },
  name: 'todo_text_search'
});

print('MongoDB initialization completed successfully');
print('Created collections: users, todos');
print('Created indexes for performance optimization');