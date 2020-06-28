package product

import (
	"strconv"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"

	v1 "github.com/caicloud/nirvana-practice/pkg/apis/meta/v1"
	api "github.com/caicloud/nirvana-practice/pkg/apis/v1alpha1"
	"github.com/caicloud/nirvana-practice/pkg/errors"
	"github.com/caicloud/nirvana-practice/pkg/tools/apis/v1alpha1"
)

type Repository interface {
	Add(product *api.Product) (*api.Product, error)
	GetAll(start, limit int, orderKey string, reverseOrder bool) []*api.Product
	Get(name string) *api.Product
	Delete(name string) error
	Update(name string, product *api.Product) error
}

type Service struct {
	products map[string]*api.Product
	lock     sync.RWMutex
}

func NewService() *Service {
	cache := make(map[string]*api.Product)
	for i := 0; i < 10; i++ {
		creationTimestamp := time.Now()
		soldTimestamp := time.Now().Add(time.Hour * 24)
		price := 22.3
		sold := true
		product := &api.Product{
			Metadata: v1.Metadata{
				UID:   uuid.NewV4().String(),
				Name:  "product" + strconv.Itoa(i),
				Alias: "产品" + strconv.Itoa(i),
				Labels: map[string]string{
					"label1": "labelTest",
				},
				Annotations: map[string]string{
					"annotations": "annotationsTest",
				},
				Description:       "这是一个普通的产品",
				CreationTimestamp: &creationTimestamp,
			},
			Spec: &api.ProductSpec{
				Category: "test",
				Price:    &price,
			},
			Status: &api.ProductStatus{
				Sold:          &sold,
				SoldTimestamp: &soldTimestamp,
			},
		}
		cache[product.Name] = product
	}
	return &Service{
		products: cache,
	}
}

func (p *Service) Add(product *api.Product) (*api.Product, error) {
	if p.exist(product.Name) {
		return nil, errors.ErrorAlreadyExist.Error(product.Name)
	}
	uid := uuid.NewV4().String()
	product.UID = uid
	p.lock.Lock()
	p.products[product.Name] = product
	p.lock.Unlock()
	return product, nil
}

func (p *Service) GetAll(start, limit int, orderKey string, reverseOrder bool) []*api.Product {
	p.lock.RLock()
	defer p.lock.RUnlock()
	products := make([]*api.Product, 0, len(p.products))
	for _, v := range p.products {
		products = append(products, v)
	}
	v1alpha1.SortByKey(products, orderKey, reverseOrder)
	if start >= len(products) {
		start = len(products) - 1
	}
	end := start + limit
	if end > len(products) {
		end = len(products)
	}
	return products[start:end]
}

func (p *Service) Get(name string) *api.Product {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if product, ok := p.products[name]; ok {
		return product
	}
	return nil
}

func (p *Service) Delete(name string) error {
	if !p.exist(name) {
		return errors.ErrorNotFound.Error(name)
	}
	p.lock.Lock()
	delete(p.products, name)
	p.lock.Unlock()
	return nil
}

func (p *Service) Update(name string, product *api.Product) error {
	if product.Name != name {
		return errors.ErrorValidationFailed.Error(name)
	}
	if !p.exist(name) {
		return errors.ErrorNotFound.Error(name)
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	p.products[name] = product
	return nil
}

func (p *Service) exist(name string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	_, ok := p.products[name]
	return ok
}