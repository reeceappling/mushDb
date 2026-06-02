import {Note} from "@/app/components/formSubcomponents/notes";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {SlantSelector} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {SporeSwabSelector} from "@/app/components/sporeSwabClient";


export function TestSporeSwabOk(){
    return new SporeSwab({
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        sale: "SALE ID",
        disposed: Date.now(),
        notes: TestNotes,
        lastUpdated: 789,
    })
}

export interface SporeSwab {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    parentType?:string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    disposed?: number
    transfersOut?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class SporeSwab {
    // Accept a single object containing the fields
    constructor(init?: Partial<SporeSwab>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "sporeSwab"
    }
}

export function SporeSwabSelectorCloseable(sp: SelectorProps<SporeSwab>) { // TODO: use
    const doSel = (val?: SporeSwab):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<SporeSwab> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Spore Swab List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "spore swab",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: SporeSwab) => void)=>{
            return <SporeSwabSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}