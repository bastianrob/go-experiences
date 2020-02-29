package reduce

import (
	"reflect"
	"testing"
)

func TestReduce(t *testing.T) {
	type Person struct {
		Name       string
		Birthplace string
	}
	type PersonGroup map[string][]string
	type SumAvg struct {
		Sum int
		Avg float32
	}

	type args struct {
		source       interface{}
		initialValue interface{}
		reducer      interface{}
	}

	sumOfInt := func(accumulator, entry, idx int) int {
		return accumulator + entry
	}

	avgOfInt := func(accumulator SumAvg, entry, idx int) SumAvg {
		sum := accumulator.Sum + entry
		return SumAvg{
			Sum: sum,
			Avg: float32(sum) / float32(idx+1),
		}
	}

	groupBirthplacesByName := func(accumulator PersonGroup, entry Person, idx int) PersonGroup {
		birthplaces, exists := accumulator[entry.Name]
		if !exists {
			birthplaces = []string{entry.Birthplace}
		} else {
			birthplaces = append(birthplaces, entry.Birthplace)
		}
		accumulator[entry.Name] = birthplaces
		return accumulator
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "Source must be an array",
			args:    args{source: "something"},
			wantErr: true,
		},
		{
			name:    "Reducer must not be nil",
			args:    args{source: []int{1, 2, 3}, reducer: nil},
			wantErr: true,
		},
		{
			name:    "Reducer must be a function",
			args:    args{source: []int{1, 2, 3}, reducer: "something"},
			wantErr: true,
		},
		{
			name: "Sum of array",
			args: args{
				source:       []int{1, 2, 3},
				initialValue: 0,
				reducer:      sumOfInt,
			},
			wantErr: false,
			want:    6,
		},
		{
			name: "Avg of array",
			args: args{
				source:       []int{1, 2, 3},
				initialValue: SumAvg{Sum: 0, Avg: 0},
				reducer:      avgOfInt,
			},
			wantErr: false,
			want: SumAvg{
				Sum: 6,
				Avg: 6 / 3,
			},
		},
		{
			name: "Group by person's name",
			args: args{
				source: []Person{
					Person{"John Doe", "Jakarta"},
					Person{"John Doe", "Depok"},
					Person{"John Doe", "Medan"},
				},
				initialValue: make(PersonGroup),
				reducer:      groupBirthplacesByName,
			},
			wantErr: false,
			want:    PersonGroup{"John Doe": []string{"Jakarta", "Depok", "Medan"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Reduce(tt.args.source, tt.args.initialValue, tt.args.reducer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reduce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}
