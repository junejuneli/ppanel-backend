package billing_test

import (
	"context"
	"fmt"
	"testing"

	dto "github.com/perfect-panel/server/internal/module/billing/contract"
	"github.com/perfect-panel/server/internal/module/billing/entity/order"
	"github.com/perfect-panel/server/internal/module/billing/internal/billingtest"
	"github.com/perfect-panel/server/internal/module/subscription/entity/subscribe"
	"github.com/perfect-panel/server/pkg/xerr"
)

func TestOneTimePlanUsesPurchaseHistoryAndFixedDuration(t *testing.T) {
	for _, status := range []uint8{order.StatusPending, order.StatusPaid, order.StatusFinished, order.StatusClosed} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			f := newFacade(t)
			buyer := f.h.User()
			plan := f.h.Plan(490, func(p *subscribe.Subscribe) {
				p.OneTime = true
				p.Quota = 1
				p.UnitTime = "Day"
				p.Discount = `[{"quantity":3,"discount":33.33333333333334}]`
			})
			method := f.h.Payment("EPay", epayConfig)
			catalog, catalogErr := f.svc.GetPortalSubscription(context.Background(), &dto.GetSubscriptionRequest{})
			if catalogErr != nil || len(catalog.List) != 1 || !catalog.List[0].OneTime {
				t.Fatalf("trial policy missing from storefront: %+v err=%v", catalog, catalogErr)
			}
			ctx := billingtest.UserContext(buyer)
			req := &dto.PurchaseOrderRequest{SubscribeId: plan.Id, Quantity: 3, Payment: method.Id}
			preview, err := f.svc.PreCreateOrder(ctx, req)
			if err != nil || preview.Amount != 490 {
				t.Fatalf("fixed trial preview=%+v err=%v", preview, err)
			}
			_, err = f.svc.PreCreateOrder(context.Background(), req)
			assertCode(t, err, xerr.InvalidAccess)
			_, err = f.svc.PreCreateOrder(ctx, &dto.PurchaseOrderRequest{SubscribeId: plan.Id, Quantity: 1, Payment: method.Id})
			assertCode(t, err, xerr.InvalidParams)
			f.h.Order(&order.Order{OrderNo: "trial-history", UserId: buyer.Id, SubscribeId: plan.Id, Type: order.TypeSubscribe, Status: status})
			_, err = f.svc.PreCreateOrder(ctx, req)
			if status == order.StatusClosed {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				assertCode(t, err, xerr.SubscribeQuotaLimit)
			}
			_, err = f.svc.Purchase(ctx, req)
			if status == order.StatusClosed {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				assertCode(t, err, xerr.SubscribeQuotaLimit)
			}
			_, err = f.svc.PreCreateOrder(ctx, &dto.PurchaseOrderRequest{SubscribeId: plan.Id, Quantity: 3, Payment: method.Id, UserSubscribeId: 99})
			assertCode(t, err, xerr.SubscribeNotAvailable)
		})
	}
}
