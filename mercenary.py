class Mercenary:
    """
    Represents a mercenary/crew member for hire.
    """
    def __init__(self, name, role, skills, salary, status='active'):
        self.name = name
        self.role = role  # e.g., pilot, fighter, trader, engineer
        self.skills = skills  # dict of skill: level
        self.salary = salary  # cost per period
        self.status = status

    def to_dict(self):
        return {
            'name': self.name,
            'role': self.role,
            'skills': self.skills,
            'salary': self.salary,
            'status': self.status,
        }

    @classmethod
    def from_dict(cls, data):
        return cls(
            name=data['name'],
            role=data['role'],
            skills=data['skills'],
            salary=data['salary'],
            status=data.get('status', 'active')
        )

    def __repr__(self):
        return f"<Mercenary {self.name} ({self.role}) skill={self.skills} salary={self.salary}>"
